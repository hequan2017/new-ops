// Package service 白泽批量作业：批量命令执行编排
// 流程：权限校验（888 全量 / 普通用户仅授权资产组内主机）→ 建批次落库 →
// 后台 goroutine 经并发池逐主机 SSH 直连执行（指纹校验/TOFU 复用 asset 通道）→
// 每主机结果与批次汇总落库；进行中批次可经 cancelRegistry 取消。
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/model/common/request"
	assetModel "github.com/hequan2017/new-ops/server/plugin/asset/model"
	assetSvc "github.com/hequan2017/new-ops/server/plugin/asset/service"
	"github.com/hequan2017/new-ops/server/plugin/job/model"
	jobRequest "github.com/hequan2017/new-ops/server/plugin/job/model/request"
)

// JobService 批量作业服务
var JobService = new(jobService)

type jobService struct{}

// cancelRegistry 进行中批次的取消函数注册表（进程内存；重启后遗留"执行中"批次视为孤儿，由人工重跑）
var cancelRegistry sync.Map // recordID(uint) -> context.CancelFunc

// 默认与上限参数
const (
	defaultConcurrency = 10
	maxConcurrency     = 50
	defaultTimeoutSec  = 30
	minTimeoutSec      = 5
	maxTimeoutSec      = 600
)

// CreateBatchExec 发起批量执行：校验 + 建批次 + 后台异步执行，立即返回批次
func (s *jobService) CreateBatchExec(req jobRequest.BatchExecReq, userID uint, isSuperAdmin bool, operator string) (*model.JobExecRecord, error) {
	// 命令解析：脚本优先，其次直接命令；变量组渲染 {{key}} 后作为批次快照
	command := req.Command
	if req.ScriptID != nil && *req.ScriptID != 0 {
		var sc model.JobScript
		if err := global.GVA_DB.First(&sc, *req.ScriptID).Error; err != nil {
			return nil, newJobErr(ErrCodeParamInvalid, fmt.Sprintf("脚本不存在(id=%d)", *req.ScriptID))
		}
		command = sc.Content
	}
	if command == "" || len(req.HostIDs) == 0 {
		return nil, newJobErr(ErrCodeParamInvalid, "命令（或脚本）与目标主机不能为空")
	}
	if req.VariableGroupID != nil && *req.VariableGroupID != 0 {
		var vg model.JobVariableGroup
		if err := global.GVA_DB.First(&vg, *req.VariableGroupID).Error; err != nil {
			return nil, newJobErr(ErrCodeParamInvalid, fmt.Sprintf("变量组不存在(id=%d)", *req.VariableGroupID))
		}
		rendered, rerr := RenderTemplate(command, vg.Variables)
		if rerr != nil {
			return nil, rerr
		}
		command = rendered
	}
	// 主机去重
	seen := make(map[uint]struct{}, len(req.HostIDs))
	hostIDs := make([]uint, 0, len(req.HostIDs))
	for _, id := range req.HostIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		hostIDs = append(hostIDs, id)
	}
	// 数据权限：非 888 仅允许授权资产组内主机
	if !isSuperAdmin {
		authorized, err := assetSvc.Service.AssetGroupService.GetUserAuthorizedHostIDs(userID)
		if err != nil {
			return nil, fmt.Errorf("查询授权主机失败: %w", err)
		}
		allow := make(map[uint]struct{}, len(authorized))
		for _, id := range authorized {
			allow[id] = struct{}{}
		}
		for _, id := range hostIDs {
			if _, ok := allow[id]; !ok {
				return nil, newJobErr(ErrCodeHostNotAuthorized, fmt.Sprintf("主机 id=%d 不在你的授权资产范围内", id))
			}
		}
	}
	if req.Concurrency <= 0 {
		req.Concurrency = defaultConcurrency
	}
	if req.Concurrency > maxConcurrency {
		req.Concurrency = maxConcurrency
	}
	if req.TimeoutSec <= 0 {
		req.TimeoutSec = defaultTimeoutSec
	}
	if req.TimeoutSec < minTimeoutSec {
		req.TimeoutSec = minTimeoutSec
	}
	if req.TimeoutSec > maxTimeoutSec {
		req.TimeoutSec = maxTimeoutSec
	}

	// 拉取主机快照（软删除/不存在的主机直接报参数错）
	hosts := make([]*assetModel.AssetHost, 0, len(hostIDs))
	for _, id := range hostIDs {
		h, err := assetSvc.Service.AssetHostService.GetAssetHost(id)
		if err != nil {
			return nil, newJobErr(ErrCodeParamInvalid, fmt.Sprintf("主机 id=%d 不存在", id))
		}
		hosts = append(hosts, h)
	}

	record := &model.JobExecRecord{
		Command:      command, // 渲染后的快照
		Concurrency:  req.Concurrency,
		TimeoutSec:   req.TimeoutSec,
		CredentialID: req.CredentialID,
		Total:        len(hosts),
		Status:       model.ExecStatusRunning,
		Operator:     operator,
		UserID:       userID,
		StartedAt:    time.Now(),
	}
	if err := global.GVA_DB.Create(record).Error; err != nil {
		return nil, err
	}
	go s.runBatch(record, hosts)
	return record, nil
}

// runBatch 批次执行主体（后台 goroutine）
func (s *jobService) runBatch(record *model.JobExecRecord, hosts []*assetModel.AssetHost) {
	ctx, cancel := context.WithCancel(context.Background())
	cancelRegistry.Store(record.ID, cancel)
	defer func() {
		cancelRegistry.Delete(record.ID)
		cancel()
	}()

	tasks := make([]PoolTask, 0, len(hosts))
	hostByID := make(map[uint]*assetModel.AssetHost, len(hosts))
	for _, h := range hosts {
		hostByID[h.ID] = h
		tasks = append(tasks, PoolTask{HostID: h.ID, Run: s.sshRunTask(ctx, h, record.CredentialID, record.Command)})
	}
	results := runPool(ctx, tasks, record.Concurrency, time.Duration(record.TimeoutSec)*time.Second)

	success, failed, canceled := 0, 0, false
	for _, r := range results {
		st := resultStatus(r)
		if st == model.ResultSuccess {
			success++
		} else {
			failed++
		}
		if r.Canceled {
			canceled = true
		}
		h := hostByID[r.HostID]
		row := &model.JobExecResult{
			RecordID: record.ID, HostID: r.HostID,
			Status: st, Output: r.Output, ElapsedMs: r.Elapsed.Milliseconds(),
		}
		if h != nil {
			row.Hostname, row.IP = h.Hostname, h.IP
		}
		if r.Err != nil {
			row.Error = r.Err.Error()
		}
		if err := global.GVA_DB.Create(row).Error; err != nil {
			global.GVA_LOG.Error(fmt.Sprintf("job 批次 %d 主机 %d 结果落库失败: %v", record.ID, r.HostID, err))
		}
	}

	now := time.Now()
	status := model.ExecStatusFinished
	if canceled {
		status = model.ExecStatusCanceled
	}
	if err := global.GVA_DB.Model(&model.JobExecRecord{}).Where("id = ?", record.ID).Updates(map[string]any{
		"status": status, "success": success, "failed": failed, "finished_at": &now,
	}).Error; err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("job 批次 %d 汇总更新失败: %v", record.ID, err))
	}
}

// sshRunTask 构造单主机 SSH 执行体：直连 + 指纹校验/TOFU + ctx 感知中断
// （跳板级联为 term 通道能力，批量执行暂直连，级联支持登记 DEV_PLAN 待办）
func (s *jobService) sshRunTask(ctx context.Context, host *assetModel.AssetHost, credOverride *uint, command string) func(context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		if host.IP == "" {
			return "", newJobErr(ErrCodeParamInvalid, fmt.Sprintf("主机 %s(id=%d) 缺少内网IP", host.Hostname, host.ID))
		}
		credID := host.CredentialID
		if credOverride != nil && *credOverride != 0 {
			credID = credOverride
		}
		if credID == nil || *credID == 0 {
			return "", newJobErr(ErrCodeNoCredential, fmt.Sprintf("主机 %s(%s) 未绑定SSH凭据且未指定统一凭据", host.Hostname, host.IP))
		}
		secret, credType, username, err := assetSvc.Service.CredCredentialService.GetPlaintext(*credID)
		if err != nil {
			return "", fmt.Errorf("凭据读取失败(id=%d): %w", *credID, err)
		}
		if credType != assetModel.CredTypeSSHPassword && credType != assetModel.CredTypeSSHKey {
			return "", newJobErr(ErrCodeNoCredential, fmt.Sprintf("凭据 id=%d 非 SSH 密码/私钥类型", *credID))
		}
		client, fp, err := assetSvc.DialSSH(host.IP, assetSvc.SSHAuth{
			Username: username, Password: secret, PrivateKey: secret,
			ExpectedFingerprint: host.SSHFP,
		})
		if err != nil {
			return "", err
		}
		defer client.Close()
		if host.SSHFP == "" {
			_ = assetSvc.Service.AssetHostService.RecordHostFingerprint(host.ID, fp, "job批量执行")
		}

		session, err := client.NewSession()
		if err != nil {
			return "", fmt.Errorf("创建会话失败: %w", err)
		}
		defer session.Close()

		type runOut struct {
			out []byte
			err error
		}
		ch := make(chan runOut, 1)
		go func() {
			out, e := session.CombinedOutput(command)
			ch <- runOut{out, e}
		}()
		select {
		case r := <-ch:
			return string(r.out), r.err
		case <-ctx.Done():
			_ = session.Close() // 中断阻塞的 CombinedOutput
			return "", ctx.Err()
		}
	}
}

// CancelBatchExec 取消进行中的批次（仅对运行中批次有效）
func (s *jobService) CancelBatchExec(recordID uint) error {
	v, ok := cancelRegistry.Load(recordID)
	if !ok {
		return newJobErr(ErrCodeBatchNotRunning, "批次不存在或已结束，无法取消")
	}
	v.(context.CancelFunc)()
	return nil
}

// GetRecordList 批次分页列表（operator 过滤：普通用户仅见自己的批次，888 全量）
func (s *jobService) GetRecordList(info request.PageInfo, userID uint, isSuperAdmin bool, status string) (list []*model.JobExecRecord, total int64, err error) {
	db := global.GVA_DB.Model(&model.JobExecRecord{})
	if !isSuperAdmin {
		db = db.Where("user_id = ?", userID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = db.Scopes(info.Paginate()).Order("id DESC").Find(&list).Error
	return list, total, err
}

// GetRecordDetail 批次详情 + 每主机结果（发起人本人或 888 可见）
func (s *jobService) GetRecordDetail(recordID uint, userID uint, isSuperAdmin bool) (*model.JobExecRecord, []*model.JobExecResult, error) {
	var record model.JobExecRecord
	if err := global.GVA_DB.First(&record, recordID).Error; err != nil {
		return nil, nil, err
	}
	if !isSuperAdmin && record.UserID != userID {
		return nil, nil, newJobErr(ErrCodeHostNotAuthorized, "只能查看自己发起的批次")
	}
	var results []*model.JobExecResult
	if err := global.GVA_DB.Where("record_id = ?", recordID).Order("id ASC").Limit(1000).Find(&results).Error; err != nil {
		return nil, nil, err
	}
	return &record, results, nil
}

// resultStatus 池结果 → 结果状态
func resultStatus(r PoolResult) string {
	switch {
	case r.Canceled:
		return model.ResultCanceled
	case r.TimedOut:
		return model.ResultTimeout
	case r.Err == nil:
		return model.ResultSuccess
	default:
		return model.ResultFailed
	}
}
