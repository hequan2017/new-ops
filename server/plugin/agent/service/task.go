// Package service 白泽 Agent 通道：任务下发与结果回传（M9 A2）
// 流程：RunTask 生成任务号 → 经连接注册表下发 task 帧 → 挂起等待 task_result（超时兜底）→
// 结果落 agent_tasks 留痕；agent 侧按任务号去重保证幂等（同号任务只执行一次）。
package service

import (
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/hequan2017/new-ops/server/global"
	model "github.com/hequan2017/new-ops/server/plugin/agent/model"
)

// agent 任务参数与留存上限
const (
	defaultTaskTimeoutSec = 60
	maxTaskTimeoutSec     = 600
	// maxTaskOutputBytes 输出留存上限（DB text 列 64KB 口径内）
	maxTaskOutputBytes = 60 * 1024
)

// agent 任务状态
const (
	TaskStatusSuccess = "成功"
	TaskStatusFailed  = "失败"
	TaskStatusTimeout = "超时"
)

// ErrCodeAgentTaskTimeout agent 执行超时（结果未在限时内回传）
const ErrCodeAgentTaskTimeout = 2005

// pendingTasks 进行中任务等待表：taskNo → 结果 chan（cap 1，回传即取走，不重放）
var pendingTasks sync.Map

// TaskResult 任务执行结果
type TaskResult struct {
	TaskNo    string `json:"taskNo"`
	ExitCode  int    `json:"exitCode"`
	Output    string `json:"output"`
	ErrMsg    string `json:"errMsg"`
	ElapsedMs int64  `json:"elapsedMs"`
}

// RunTask 向资产上的在线 agent 下发命令并同步等待结果（超时兜底=执行限时+2s 传输余量）
func (s *AgentService) RunTask(operator string, assetID uint, command string, timeoutSec int) (*TaskResult, error) {
	if assetID == 0 || strings.TrimSpace(command) == "" {
		return nil, newAgentErr(ErrCodeAgentParamInvalid, "assetId 与 command 必填")
	}
	if timeoutSec <= 0 {
		timeoutSec = defaultTaskTimeoutSec
	}
	if timeoutSec > maxTaskTimeoutSec {
		timeoutSec = maxTaskTimeoutSec
	}
	taskNo := uuid.NewString()
	ch := make(chan agentFrame, 1)
	pendingTasks.Store(taskNo, ch)
	defer pendingTasks.Delete(taskNo)

	start := time.Now()
	if err := WriteToAgent(assetID, agentFrame{
		Type: "task", TS: time.Now().Unix(), TaskNo: taskNo, Payload: command, TaskTimeout: timeoutSec,
	}); err != nil {
		return nil, err
	}

	timer := time.NewTimer(time.Duration(timeoutSec)*time.Second + 2*time.Second)
	defer timer.Stop()
	select {
	case f := <-ch:
		r := &TaskResult{
			TaskNo: taskNo, ExitCode: f.ExitCode, Output: f.Output, ErrMsg: f.Msg,
			ElapsedMs: time.Since(start).Milliseconds(),
		}
		s.persistTask(operator, assetID, command, taskStatusOf(r.ExitCode, r.ErrMsg), r)
		return r, nil
	case <-timer.C:
		r := &TaskResult{TaskNo: taskNo, ErrMsg: "agent 执行超时", ElapsedMs: time.Since(start).Milliseconds()}
		s.persistTask(operator, assetID, command, TaskStatusTimeout, r)
		return nil, newAgentErr(ErrCodeAgentTaskTimeout, "agent 执行超时（结果未在限时内回传）")
	}
}

// ListAgentTasks 任务留痕分页（assetId 可选过滤）
func (s *AgentService) ListAgentTasks(page, pageSize int, assetID uint) ([]model.AgentTask, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	db := global.GVA_DB.Model(&model.AgentTask{})
	if assetID > 0 {
		db = db.Where("asset_id = ?", assetID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.AgentTask
	err := db.Order("id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error
	return list, total, err
}

// resolveTaskResult task_result 帧路由到等待方（读循环调用；无等待方/重复回传直接丢弃）
func resolveTaskResult(taskNo string, f agentFrame) bool {
	if taskNo == "" {
		return false
	}
	v, ok := pendingTasks.Load(taskNo)
	if !ok {
		return false
	}
	pendingTasks.Delete(taskNo)
	ch := v.(chan agentFrame)
	select {
	case ch <- f:
		return true
	default:
		return false
	}
}

// taskStatusOf 结果 → 状态（纯逻辑）：错误文案含"超时"→超时；非零退出码或错误信息→失败
func taskStatusOf(exitCode int, errMsg string) string {
	switch {
	case strings.Contains(errMsg, "超时"):
		return TaskStatusTimeout
	case errMsg != "" || exitCode != 0:
		return TaskStatusFailed
	default:
		return TaskStatusSuccess
	}
}

// persistTask 任务留痕（best-effort：留痕失败仅记日志，不阻断调用方）
func (s *AgentService) persistTask(operator string, assetID uint, command, status string, r *TaskResult) {
	if global.GVA_DB == nil {
		return
	}
	now := time.Now()
	row := model.AgentTask{
		TaskNo: r.TaskNo, AssetID: assetID, Type: "exec", Command: command, Status: status,
		ExitCode: r.ExitCode, Output: truncateBytes(r.Output, maxTaskOutputBytes), ErrMsg: r.ErrMsg,
		ElapsedMs: r.ElapsedMs, Operator: operator, FinishedAt: &now,
	}
	if err := global.GVA_DB.Create(&row).Error; err != nil {
		global.GVA_LOG.Error("agent 任务留痕失败: " + err.Error())
	}
}

// truncateBytes 按 UTF-8 字符边界截断到 max 字节内（纯逻辑）
func truncateBytes(s string, max int) string {
	if len(s) <= max {
		return s
	}
	i := max
	for i > 0 && !utf8.RuneStart(s[i]) {
		i--
	}
	return s[:i]
}
