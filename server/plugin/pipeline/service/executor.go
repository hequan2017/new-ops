// Package service 白泽流水线执行器（M3 场26）
// 设计对齐 DEV_PLAN 关键决策 3 与 M3「发布目标绑定资产与凭据」：
// shell 步骤一律经 SSH 在绑定资产上执行（复用指纹校验/TOFU 通道，无本机进程执行面）；
// 状态机：等待中 → 执行中 →（等待审批）→ 成功 | 失败 | 已取消；
// 触发即对定义做快照（参数 {{key}} 已替换），构建期间改定义不影响本次构建。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hequan2017/new-ops/server/global"
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
			Steps:           make([]snapshotStep, 0, len(st.Steps)),
		}
		for _, sp := range st.Steps {
			ss.Steps = append(ss.Steps, snapshotStep{
				Name:         sp.Name,
				Type:         sp.Type,
				HostID:       sp.HostID,
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
	if err := global.GVA_DB.Create(build).Error; err != nil {
		return nil, err
	}
	go s.run(build.ID)
	return build, nil
}

// run 执行主体（后台 goroutine）：状态机推进 + 审批 gate + 串行阶段
func (s *PipelineBuildService) run(buildID uint) {
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
		for _, sp := range st.Steps {
			if err := s.runStep(ctx, buildID, st.Name, sp); err != nil {
				s.appendLog(buildID, st.Name, sp.Name, model.LogSystem, fmt.Sprintf("步骤失败: %v", err))
				if st.ContinueOnError {
					s.appendLog(buildID, st.Name, "", model.LogSystem, "阶段配置失败继续，跳过该错误")
					continue
				}
				failed = true
				break
			}
		}
	}
	if failed {
		s.finish(buildID, model.BuildFailed)
		return
	}
	s.finish(buildID, model.BuildSuccess)
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

// runShellStep 经 SSH 在绑定资产上执行 shell（指纹校验/TOFU，取消即中断会话）
func (s *PipelineBuildService) runShellStep(ctx context.Context, buildID uint, stageName string, sp snapshotStep) error {
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
	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	session.Stderr = session.Stdout
	if err := session.Start(sp.ShellContent); err != nil {
		return fmt.Errorf("启动命令失败: %w", err)
	}
	type result struct {
		out string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		b, _ := io.ReadAll(stdout) // 读错误以 Wait 退出码为准（EOF 属正常收尾）
		waitErr := session.Wait()
		ch <- result{string(b), waitErr}
	}()
	select {
	case r := <-ch:
		if out := strings.TrimSpace(r.out); out != "" {
			s.appendLog(buildID, stageName, sp.Name, model.LogStdout, out)
		}
		return r.err
	case <-ctx.Done():
		_ = session.Close() // 中断远端命令
		<-ch
		return ctx.Err()
	}
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
	snippet := make([]byte, 1024)
	n, _ := resp.Body.Read(snippet)
	s.appendLog(buildID, stageName, sp.Name, model.LogSystem,
		fmt.Sprintf("HTTP %d %s %s", resp.StatusCode, http.StatusText(resp.StatusCode), strings.TrimSpace(string(snippet[:n]))))
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

// ApproveBuild 审批放行（仅等待审批状态有效）
func (s *PipelineBuildService) ApproveBuild(buildID uint, operator string) error {
	v, ok := buildRegistry.Load(buildID)
	if !ok {
		return newPlErr(ErrCodePlNotFound, "构建不存在或已结束")
	}
	rt := v.(*buildRuntime)
	select {
	case rt.approveCh <- struct{}{}:
		s.appendLog(buildID, "", "", model.LogSystem, "人工审批放行（"+operator+"）")
		return nil
	default:
		return newPlErr(ErrCodePlStructInvalid, "构建不在等待审批状态或已放行")
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

// finish 收档（成功/失败/取消统一结束时间）
func (s *PipelineBuildService) finish(buildID uint, status string) {
	now := time.Now()
	s.updateStatus(buildID, status, map[string]any{"finished_at": &now})
	s.appendLog(buildID, "", "", model.LogSystem, "构建结束："+status)
}

// appendLog 落一条构建日志（长内容截断 16KB）
func (s *PipelineBuildService) appendLog(buildID uint, stageName, stepName, level, content string) {
	if len(content) > 16*1024 {
		content = content[:16*1024] + "\n…(截断)"
	}
	if err := global.GVA_DB.Create(&model.BuildLog{
		BuildID: buildID, StageName: stageName, StepName: stepName, Level: level, Content: content,
	}).Error; err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("构建 %d 日志落库失败: %v", buildID, err))
	}
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

// GetBuildLogs 构建日志（升序分页）
func (s *PipelineBuildService) GetBuildLogs(buildID uint, page, pageSize int) (list []*model.BuildLog, total int64, err error) {
	db := global.GVA_DB.Model(&model.BuildLog{}).Where("build_id = ?", buildID)
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
