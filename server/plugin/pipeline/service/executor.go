// Package service 白泽流水线执行器（M3 场26）
// 设计要点（M3「发布目标绑定资产与凭据」）：
// shell 步骤一律经 SSH 在绑定资产上执行（复用指纹校验/TOFU 通道，无本机进程执行面）；
// 状态机：等待中 → 执行中 →（等待审批）→ 成功 | 失败 | 已取消；
// 触发即对定义做快照（参数 {{key}} 已替换），构建期间改定义不影响本次构建。
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/hequan2017/new-ops/server/global"
	agentSvc "github.com/hequan2017/new-ops/server/plugin/agent/service"
	assetModel "github.com/hequan2017/new-ops/server/plugin/asset/model"
	assetSvc "github.com/hequan2017/new-ops/server/plugin/asset/service"
	"github.com/hequan2017/new-ops/server/plugin/pipeline/model"
)

// PipelineBuildService 构建服务
type PipelineBuildService struct{}

// pipelineSvcForBuild 定义服务指针（CreateBuild 复用详情预载）
var pipelineSvcForBuild = new(PipelineService)

// buildRuntime 进行中构建的运行时（取消与审批用）
type buildRuntime struct {
	cancel    context.CancelFunc
	approveCh chan struct{} // 审批 gate 放行信号
}

var buildRegistry sync.Map // buildID(uint) -> *buildRuntime

// snapshotStep / snapshotStage / buildSnapshot 定义快照（变量替换后）
type snapshotStep struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	HostID       uint   `json:"hostId"`
	ExecBackend  string `json:"execBackend,omitempty"` // ssh/agent（M9 A3）
	ShellContent string `json:"shellContent,omitempty"`
	HTTPMethod   string `json:"httpMethod,omitempty"`
	HTTPURL      string `json:"httpUrl,omitempty"`
	HTTPBody     string `json:"httpBody,omitempty"`
	TimeoutSec   int    `json:"timeoutSec"`
}

type snapshotStage struct {
	Name            string         `json:"name"`
	Approval        bool           `json:"approval"`
	ContinueOnError bool           `json:"continueOnError"`
	Parallel        bool           `json:"parallel"`
	Steps           []snapshotStep `json:"steps"`
}

type buildSnapshot struct {
	Stages []snapshotStage `json:"stages"`
}

// renderVars 参数替换：{{key}} → 值（未知占位保持原样）
func renderVars(content string, params map[string]string) string {
	for k, v := range params {
		if k == "" {
			continue
		}
		content = strings.ReplaceAll(content, "{{"+k+"}}", v)
	}
	return content
}

// CreateBuild 触发构建：校验 → 快照渲染 → 落库 → 后台执行，立即返回构建记录
func (s *PipelineBuildService) CreateBuild(pipelineID uint, params map[string]string, operator string, userID uint) (*model.PipelineBuild, error) {
	pl, err := pipelineSvcForBuild.GetPipelineDetail(pipelineID)
	if err != nil {
		return nil, err
	}
	if !pl.Enabled {
		return nil, newPlErr(ErrCodePlStructInvalid, "流水线已停用，无法触发")
	}
	if params == nil {
		params = map[string]string{}
	}
	snap := buildSnapshot{Stages: make([]snapshotStage, 0, len(pl.Stages))}
	for _, st := range pl.Stages {
		ss := snapshotStage{
			Name:            st.Name,
			Approval:        st.Approval,
			ContinueOnError: st.ContinueOnError,
			Parallel:        st.Parallel,
			Steps:           make([]snapshotStep, 0, len(st.Steps)),
		}
		for _, sp := range st.Steps {
			ss.Steps = append(ss.Steps, snapshotStep{
				Name:         sp.Name,
				Type:         sp.Type,
				HostID:       sp.HostID,
				ExecBackend:  sp.ExecBackend,
				ShellContent: renderVars(sp.ShellContent, params),
				HTTPMethod:   sp.HTTPMethod,
				HTTPURL:      renderVars(sp.HTTPURL, params),
				HTTPBody:     renderVars(sp.HTTPBody, params),
				TimeoutSec:   sp.TimeoutSec,
			})
		}
		snap.Stages = append(snap.Stages, ss)
	}
	snapJSON, jErr := json.Marshal(snap)
	if jErr != nil {
		return nil, fmt.Errorf("生成快照失败: %w", jErr)
	}

	var buildNo int
	if err := global.GVA_DB.Model(&model.PipelineBuild{}).
		Where("pipeline_id = ?", pipelineID).
		Select("COALESCE(MAX(build_no),0)+1").Scan(&buildNo).Error; err != nil {
		return nil, err
	}
	paramsJSON, _ := json.Marshal(params)
	build := &model.PipelineBuild{
		PipelineID:   pipelineID,
		PipelineName: pl.Name,
		BuildNo:      buildNo,
		Status:       model.BuildPending,
		Params:       string(paramsJSON),
		Snapshot:     string(snapJSON),
		Operator:     operator,
		UserID:       userID,
	}
	// 取号为先读 MAX 后写，并发触发会撞号：模型有 (pipeline_id, build_no) 唯一索引兜底，
	// 撞号时 Create 报错，重取一次（webhook/手动/cron 并发窗口内概率极低，重试一次足够）
	if err := global.GVA_DB.Create(build).Error; err != nil {
		var retryNo int
		if rErr := global.GVA_DB.Model(&model.PipelineBuild{}).
			Where("pipeline_id = ?", pipelineID).
			Select("COALESCE(MAX(build_no),0)+1").Scan(&retryNo).Error; rErr == nil && retryNo > buildNo {
			build.BuildNo = retryNo
			if err2 := global.GVA_DB.Create(build).Error; err2 != nil {
				return nil, err2
			}
		} else {
			return nil, err
		}
	}
	go s.run(build.ID)
	return build, nil
}

// run 执行主体（后台 goroutine）：状态机推进 + 审批 gate + 串行阶段
func (s *PipelineBuildService) run(buildID uint) {
	// 后台执行 goroutine 无 cron 级 panic 恢复兜底：不 recover 会打死整个 server 进程，
	// 且构建永久停留"执行中"（审批/取消入口全部失效）——收档为失败并留痕
	defer func() {
		if r := recover(); r != nil {
			global.GVA_LOG.Error(fmt.Sprintf("构建 %d 执行 panic: %v\n%s", buildID, r, debug.Stack()))
			s.appendLog(buildID, "", "", model.LogSystem, fmt.Sprintf("构建执行异常崩溃: %v", r))
			s.finish(buildID, model.BuildFailed)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	rt := &buildRuntime{cancel: cancel, approveCh: make(chan struct{}, 1)}
	buildRegistry.Store(buildID, rt)
	defer func() {
		buildRegistry.Delete(buildID)
		cancel()
	}()

	var build model.PipelineBuild
	if err := global.GVA_DB.First(&build, buildID).Error; err != nil {
		return
	}
	if build.Status == model.BuildCanceled { // 极快外部取消
		return
	}
	now := time.Now()
	s.updateStatus(buildID, model.BuildRunning, map[string]any{"started_at": &now})

	var snap buildSnapshot
	if err := json.Unmarshal([]byte(build.Snapshot), &snap); err != nil {
		s.appendLog(buildID, "", "", model.LogSystem, "快照解析失败: "+err.Error())
		s.finish(buildID, model.BuildFailed)
		return
	}

	s.appendLog(buildID, "", "", model.LogSystem,
		fmt.Sprintf("构建 #%d 启动（触发人 %s），共 %d 个阶段", build.BuildNo, build.Operator, len(snap.Stages)))
	failed := false
	for si, st := range snap.Stages {
		if failed {
			break
		}
		if st.Approval {
			s.appendLog(buildID, st.Name, "", model.LogSystem, "阶段需要人工审批，等待放行…")
			s.updateStatus(buildID, model.BuildAwaiting, nil)
			select {
			case <-rt.approveCh:
				s.appendLog(buildID, st.Name, "", model.LogSystem, "审批放行，继续执行")
				s.updateStatus(buildID, model.BuildRunning, nil)
			case <-ctx.Done():
				s.appendLog(buildID, st.Name, "", model.LogSystem, "构建被取消")
				s.finish(buildID, model.BuildCanceled)
				return
			}
		}
		s.appendLog(buildID, st.Name, "", model.LogSystem,
			fmt.Sprintf("开始阶段 %d/%d「%s」", si+1, len(snap.Stages), st.Name))
		if err := s.runStageSteps(ctx, buildID, &st); err != nil {
			if ctx.Err() != nil {
				return // 用户取消：CancelBuild 已收档为已取消
			}
			failed = true
			break
		}
	}
	if failed {
		s.finish(buildID, model.BuildFailed)
		return
	}
	s.finish(buildID, model.BuildSuccess)
}

// maxParallelSteps 阶段内步骤并发上限
const maxParallelSteps = 10

// runStageSteps 执行阶段内步骤：Parallel 时并发（信号量上限，首败快速中断兄弟步骤），
// 否则串行。continueOnError=true 时吞掉失败（记日志）返回 nil。
func (s *PipelineBuildService) runStageSteps(ctx context.Context, buildID uint, st *snapshotStage) error {
	run1 := func(stepCtx context.Context, sp snapshotStep) error {
		if err := s.runStep(stepCtx, buildID, st.Name, sp); err != nil {
			s.appendLog(buildID, st.Name, sp.Name, model.LogSystem, fmt.Sprintf("步骤失败: %v", err))
			return err
		}
		return nil
	}
	if !st.Parallel || len(st.Steps) <= 1 {
		for _, sp := range st.Steps {
			if err := run1(ctx, sp); err != nil {
				if st.ContinueOnError {
					s.appendLog(buildID, st.Name, "", model.LogSystem, "阶段配置失败继续，跳过该错误")
					continue
				}
				return err
			}
		}
		return nil
	}
	s.appendLog(buildID, st.Name, "", model.LogSystem,
		fmt.Sprintf("阶段「%s」步骤并发执行（%d 步，上限 %d）", st.Name, len(st.Steps), maxParallelSteps))
	stepCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	errCh := make(chan error, len(st.Steps))
	sem := make(chan struct{}, maxParallelSteps)
	var wg sync.WaitGroup
	for _, sp := range st.Steps {
		wg.Add(1)
		go func(sp snapshotStep) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-stepCtx.Done():
				errCh <- stepCtx.Err()
				return
			}
			if err := run1(stepCtx, sp); err != nil {
				errCh <- err
				if !st.ContinueOnError {
					cancel() // 快速失败：中断兄弟步骤
				}
				return
			}
			errCh <- nil
		}(sp)
	}
	wg.Wait()
	close(errCh)
	var firstErr error
	for e := range errCh {
		if e != nil && !errors.Is(e, context.Canceled) {
			firstErr = e // 兄弟步骤因快速失败被取消的错误不计
		}
	}
	if firstErr != nil {
		if st.ContinueOnError {
			s.appendLog(buildID, st.Name, "", model.LogSystem, "阶段配置失败继续，跳过并发步骤失败")
			return nil
		}
		return firstErr
	}
	select {
	case <-ctx.Done():
		return ctx.Err() // 上层取消（已由 CancelBuild 收档）
	default:
	}
	return nil
}

// runStep 执行单步骤：shell（SSH 到绑定资产）或 http
func (s *PipelineBuildService) runStep(ctx context.Context, buildID uint, stageName string, sp snapshotStep) error {
	s.appendLog(buildID, stageName, sp.Name, model.LogSystem,
		fmt.Sprintf("执行步骤「%s」（%s）", sp.Name, sp.Type))
	stepCtx := ctx
	var stepCancel context.CancelFunc
	if sp.TimeoutSec > 0 {
		stepCtx, stepCancel = context.WithTimeout(ctx, time.Duration(sp.TimeoutSec)*time.Second)
		defer stepCancel()
	}
	switch sp.Type {
	case model.StepTypeShell:
		return s.runShellStep(stepCtx, buildID, stageName, sp)
	case model.StepTypeHTTP:
		return s.runHTTPStep(stepCtx, buildID, stageName, sp)
	default:
		return fmt.Errorf("未知步骤类型: %s", sp.Type)
	}
}

// 执行后端（M9 A3：SSH 直连 / Agent 通道）
const (
	StepBackendSSH   = "ssh"
	StepBackendAgent = "agent"
)

// stepBackend 步骤执行后端归一（纯逻辑）：空=ssh；仅支持 ssh/agent
func stepBackend(s string) (string, error) {
	switch s {
	case "", StepBackendSSH:
		return StepBackendSSH, nil
	case StepBackendAgent:
		return StepBackendAgent, nil
	default:
		return "", fmt.Errorf("执行后端仅支持 ssh/agent（当前 %q）", s)
	}
}

// workspaceFor 构建工作空间：按构建隔离（POSIX 路径；构建产物留存 /tmp 供排障，不自动清理）
func workspaceFor(buildID uint) string {
	return fmt.Sprintf("/tmp/baize-pipeline/build-%d", buildID)
}

// pipelineShellCommand 工作空间包装命令（纯逻辑）：建目录+进入后以 POSIX 子Shell执行原内容，
// 退出码=子Shell退出码；圆括号子Shell在 bash/dash 均合法（agent 侧 /bin/sh 为 dash）
func pipelineShellCommand(workspace, content string) string {
	return fmt.Sprintf("mkdir -p %q && cd %q && (\n%s\n)", workspace, workspace, content)
}

// maxShellOutputBytes SSH 步骤输出采集上限（与 job 批量执行/agent 通道 60KB 口径一致）
const maxShellOutputBytes = 60 * 1024

// boundedShellWriter 有界输出缓冲（head-only）：cat 大文件/高频输出命令在超时窗口内可产出
// 远超留存量（日志截 16KB）的数据——原 io.ReadAll(stdout) 全量缓冲会在采集端先打爆内存
type boundedShellWriter struct {
	buf []byte
	max int
}

func (w *boundedShellWriter) Write(p []byte) (int, error) {
	room := w.max - len(w.buf)
	if room <= 0 {
		return len(p), nil
	}
	if len(p) > room {
		p = p[:room]
	}
	w.buf = append(w.buf, p...)
	return len(p), nil
}

func (w *boundedShellWriter) Bytes() []byte { return w.buf }

// runShellStep 在绑定资产上执行 shell（M9 A3 双后端：SSH 直连经指纹校验/TOFU；Agent 通道经 WS 下发）。
// 两后端统一以「构建工作空间」包装执行（mkdir+cd+子Shell），实现工作空间按构建隔离。
func (s *PipelineBuildService) runShellStep(ctx context.Context, buildID uint, stageName string, sp snapshotStep) error {
	backend, berr := stepBackend(sp.ExecBackend)
	if berr != nil {
		return berr
	}
	command := pipelineShellCommand(workspaceFor(buildID), sp.ShellContent)
	if backend == StepBackendAgent {
		return s.runShellStepViaAgent(ctx, buildID, stageName, sp, command)
	}
	host, err := assetSvc.Service.AssetHostService.GetAssetHost(sp.HostID)
	if err != nil {
		return fmt.Errorf("目标主机不存在(id=%d)", sp.HostID)
	}
	if host.IP == "" {
		return fmt.Errorf("目标主机 %s 缺少内网IP", host.Hostname)
	}
	if host.CredentialID == nil || *host.CredentialID == 0 {
		return fmt.Errorf("目标主机 %s 未绑定SSH凭据", host.Hostname)
	}
	secret, credType, username, err := assetSvc.Service.CredCredentialService.GetPlaintext(*host.CredentialID)
	if err != nil {
		return fmt.Errorf("凭据读取失败(id=%d): %w", *host.CredentialID, err)
	}
	if credType != assetModel.CredTypeSSHPassword && credType != assetModel.CredTypeSSHKey {
		return fmt.Errorf("凭据 id=%d 非 SSH 类型", *host.CredentialID)
	}
	client, fp, err := assetSvc.DialSSH(host.IP, assetSvc.SSHAuth{
		Username: username, Password: secret, PrivateKey: secret,
		ExpectedFingerprint: host.SSHFP,
	})
	if err != nil {
		return err
	}
	defer client.Close()
	if host.SSHFP == "" {
		_ = assetSvc.Service.AssetHostService.RecordHostFingerprint(host.ID, fp, "流水线构建")
	}
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()
	// 有界输出采集（stdout/stderr 合流），与 job 批量执行同口径
	out := &boundedShellWriter{max: maxShellOutputBytes}
	session.Stdout = out
	session.Stderr = out
	type result struct {
		err error
	}
	ch := make(chan result, 1)
	go func() {
		ch <- result{session.Run(command)}
	}()
	select {
	case r := <-ch:
		if o := strings.TrimSpace(string(out.Bytes())); o != "" {
			s.appendLog(buildID, stageName, sp.Name, model.LogStdout, o)
		}
		return r.err
	case <-ctx.Done():
		_ = session.Close() // 中断远端命令
		<-ch
		return ctx.Err()
	}
}

// runShellStepViaAgent 经 Agent 通道执行 shell 步骤（M9 A3）：无需 SSH 凭据，agent 未在线/执行超时即步骤失败；
// ctx 感知（原签名不含 ctx，取消与步骤超时均无法中断等待，与 job 的 agentRunTask 不一致）——取消/超时即返回，
// agent 侧任务靠任务号幂等跑完即弃，迟到结果不再采纳；输出上限 60KB 与通道口径一致
func (s *PipelineBuildService) runShellStepViaAgent(ctx context.Context, buildID uint, stageName string, sp snapshotStep, command string) error {
	type agentOut struct {
		result *agentSvc.TaskResult
		err    error
	}
	ch := make(chan agentOut, 1)
	go func() {
		r, err := agentSvc.AgentSvc.RunTask(fmt.Sprintf("流水线构建#%d", buildID), sp.HostID, command, sp.TimeoutSec)
		ch <- agentOut{result: r, err: err}
	}()
	var r *agentSvc.TaskResult
	select {
	case o := <-ch:
		if o.err != nil {
			return o.err
		}
		r = o.result
	case <-ctx.Done():
		return ctx.Err()
	}
	if out := strings.TrimSpace(r.Output); out != "" {
		s.appendLog(buildID, stageName, sp.Name, model.LogStdout, out)
	}
	if r.ErrMsg != "" {
		return fmt.Errorf("%s", r.ErrMsg)
	}
	if r.ExitCode != 0 {
		return fmt.Errorf("exit status %d", r.ExitCode)
	}
	return nil
}

// runHTTPStep 发起 HTTP 请求（4xx/5xx 视为失败）
func (s *PipelineBuildService) runHTTPStep(ctx context.Context, buildID uint, stageName string, sp snapshotStep) error {
	method := sp.HTTPMethod
	if method == "" {
		method = http.MethodGet
	}
	var body io.Reader
	if sp.HTTPBody != "" && method != http.MethodGet && method != http.MethodHead {
		body = strings.NewReader(sp.HTTPBody)
	}
	req, err := http.NewRequestWithContext(ctx, method, sp.HTTPURL, body)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	s.appendLog(buildID, stageName, sp.Name, model.LogSystem,
		fmt.Sprintf("HTTP %d %s %s", resp.StatusCode, http.StatusText(resp.StatusCode), strings.TrimSpace(readBodySnippet(resp.Body, 1024))))
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

// ApproveBuild 审批放行（仅等待审批状态有效；先查 DB 状态防执行长阶段时信号提前入队、
// 被下一道审批 gate 瞬间消费放行）
func (s *PipelineBuildService) ApproveBuild(buildID uint, operator string) error {
	v, ok := buildRegistry.Load(buildID)
	if !ok {
		return newPlErr(ErrCodePlNotFound, "构建不存在或已结束")
	}
	rt := v.(*buildRuntime)
	var build model.PipelineBuild
	if err := global.GVA_DB.Select("status").First(&build, buildID).Error; err != nil {
		return newPlErr(ErrCodePlNotFound, "构建不存在或已结束")
	}
	if build.Status != model.BuildAwaiting {
		return newPlErr(ErrCodePlStructInvalid, "构建不在等待审批状态")
	}
	select {
	case rt.approveCh <- struct{}{}:
		s.appendLog(buildID, "", "", model.LogSystem, "人工审批放行（"+operator+"）")
		return nil
	default:
		return newPlErr(ErrCodePlStructInvalid, "已放行，请勿重复操作")
	}
}

// CancelBuild 取消构建（等待审批/执行中均可）
func (s *PipelineBuildService) CancelBuild(buildID uint, operator string) error {
	var build model.PipelineBuild
	if err := global.GVA_DB.First(&build, buildID).Error; err != nil {
		return newPlErr(ErrCodePlNotFound, "构建不存在")
	}
	switch build.Status {
	case model.BuildPending, model.BuildAwaiting, model.BuildRunning:
	default:
		return newPlErr(ErrCodePlStructInvalid, "构建已结束，无法取消")
	}
	if v, ok := buildRegistry.Load(buildID); ok {
		v.(*buildRuntime).cancel()
	}
	s.appendLog(buildID, "", "", model.LogSystem, "构建被取消（"+operator+"）")
	s.finish(buildID, model.BuildCanceled)
	return nil
}

// updateStatus 更新构建状态
func (s *PipelineBuildService) updateStatus(buildID uint, status string, extra map[string]any) {
	updates := map[string]any{"status": status}
	for k, v := range extra {
		updates[k] = v
	}
	global.GVA_DB.Model(&model.PipelineBuild{}).Where("id = ?", buildID).Updates(updates)
}

// finish 收档（成功/失败/取消统一结束时间；先落"构建结束"日志再写终态，
// 保证 SSE 侧按终态收流前能读到这条日志——原反序会丢尾部日志行）
func (s *PipelineBuildService) finish(buildID uint, status string) {
	now := time.Now()
	s.appendLog(buildID, "", "", model.LogSystem, "构建结束："+status)
	s.updateStatus(buildID, status, map[string]any{"finished_at": &now})
}

// appendLog 落一条构建日志（长内容按 UTF-8 字符边界截断 16KB）
func (s *PipelineBuildService) appendLog(buildID uint, stageName, stepName, level, content string) {
	if len(content) > 16*1024 {
		content = truncateLogRunes(content, 16*1024)
	}
	if err := global.GVA_DB.Create(&model.BuildLog{
		BuildID: buildID, StageName: stageName, StepName: stepName, Level: level, Content: content,
	}).Error; err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("构建 %d 日志落库失败: %v", buildID, err))
	}
}

// truncateLogRunes 按 UTF-8 字符边界截断（纯逻辑；字节硬切会切出非法 UTF-8 存库乱码）
func truncateLogRunes(content string, max int) string {
	if len(content) <= max {
		return content
	}
	i := max
	for i > 0 && !utf8.RuneStart(content[i]) {
		i--
	}
	return content[:i] + "\n…(截断)"
}

// readBodySnippet 完整读取响应体前 max 字节（单次 Read 允许短读，会丢日志快照）
func readBodySnippet(r io.Reader, max int) string {
	if r == nil {
		return ""
	}
	b, _ := io.ReadAll(io.LimitReader(r, int64(max)))
	return string(b)
}

// GetBuildList 构建分页列表（pipelineID/status 过滤；普通用户仅本人）
func (s *PipelineBuildService) GetBuildList(page, pageSize int, pipelineID *uint, status, username string, userID uint, isSuperAdmin bool) (list []*model.PipelineBuild, total int64, err error) {
	db := global.GVA_DB.Model(&model.PipelineBuild{})
	if pipelineID != nil && *pipelineID > 0 {
		db = db.Where("pipeline_id = ?", *pipelineID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if !isSuperAdmin {
		db = db.Where("user_id = ?", userID)
	}
	if username != "" {
		db = db.Where("operator LIKE ?", "%"+username+"%")
	}
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	err = db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// CheckBuildOwner 构建属主校验（日志读取等按 ID 直查的入口用）：
// 888 放行；其他用户仅可访问自己的构建——原日志接口只做角色级 casbin，
// 9528 可枚举 buildId 读他人构建的全部 shell 输出（列表接口有 user_id 过滤，此处补齐同口径）
func (s *PipelineBuildService) CheckBuildOwner(buildID uint, userID uint, isSuperAdmin bool) error {
	if isSuperAdmin {
		return nil
	}
	var count int64
	global.GVA_DB.Model(&model.PipelineBuild{}).Where("id = ? AND user_id = ?", buildID, userID).Count(&count)
	if count == 0 {
		return newPlErr(ErrCodePlNotFound, "构建不存在或无权访问")
	}
	return nil
}

// GetBuildLogs 构建日志（升序分页）
func (s *PipelineBuildService) GetBuildLogs(buildID uint, page, pageSize int) (list []*model.BuildLog, total int64, err error) {	db := global.GVA_DB.Model(&model.BuildLog{}).Where("build_id = ?", buildID)
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 100
	}
	err = db.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetBuildLogsAfter 增量拉取：构建日志 id > afterID 的记录（SSE 轮询底座）
func (s *PipelineBuildService) GetBuildLogsAfter(buildID uint, afterID uint) (list []*model.BuildLog, err error) {
	err = global.GVA_DB.Where("build_id = ? AND id > ?", buildID, afterID).
		Order("id ASC").Limit(200).Find(&list).Error
	return list, err
}

// GetBuildStatus 构建状态快照（SSE 结束判定用）
func (s *PipelineBuildService) GetBuildStatus(buildID uint) (string, error) {
	var build model.PipelineBuild
	if err := global.GVA_DB.Select("status").First(&build, buildID).Error; err != nil {
		return "", err
	}
	return build.Status, nil
}

// RestartBuild 复用历史参数重跑：以原构建的参数对当前定义重新触发（新构建新快照）
func (s *PipelineBuildService) RestartBuild(buildID uint, operator string, userID uint) (*model.PipelineBuild, error) {
	var old model.PipelineBuild
	if err := global.GVA_DB.First(&old, buildID).Error; err != nil {
		return nil, newPlErr(ErrCodePlNotFound, "构建不存在")
	}
	var params map[string]string
	if old.Params != "" {
		if err := json.Unmarshal([]byte(old.Params), &params); err != nil {
			params = map[string]string{}
		}
	}
	build, err := s.CreateBuild(old.PipelineID, params, operator, userID)
	if err != nil {
		return nil, err
	}
	s.appendLog(build.ID, "", "", model.LogSystem, fmt.Sprintf("复用构建 #%d 参数重跑（%s）", old.BuildNo, operator))
	return build, nil
}

// WebhookTrigger webhook 触发：令牌即凭据（public 端点无登录态），未开启/令牌不符一律 404 语义
func (s *PipelineBuildService) WebhookTrigger(token string, params map[string]string) (*model.PipelineBuild, error) {
	if token == "" {
		return nil, newPlErr(ErrCodePlNotFound, "无效的 webhook 令牌")
	}
	var pl model.Pipeline
	if err := global.GVA_DB.Where("webhook_token = ? AND webhook_enabled = ? AND enabled = ?",
		token, true, true).First(&pl).Error; err != nil {
		return nil, newPlErr(ErrCodePlNotFound, "无效的 webhook 令牌")
	}
	return s.CreateBuild(pl.ID, params, "webhook", 0)
}
