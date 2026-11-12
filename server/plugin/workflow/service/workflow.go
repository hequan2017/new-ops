// Package service 白泽工单引擎（M7）：定义校验 + 实例流转
// workflow 插件错误码段：1900-1999
package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/workflow/model"
	"github.com/hequan2017/new-ops/server/plugin/workflow/repo"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码
const (
	ErrCodeWfStore         = 1900
	ErrCodeWfNameRequired  = 1901
	ErrCodeWfStructInvalid = 1902
	ErrCodeWfNotFound      = 1903
	ErrCodeWfActionInvalid = 1904
)

// WorkflowService 工单服务（依赖注入：db 为 nil 时惰性回退底座全局库——
// 包级单例构造早于底座建库，禁止在构造期捕获 GVA_DB；trigger 为 nil 时用默认流水线触发器）
type WorkflowService struct {
	db      *gorm.DB
	repo    repo.Repo
	trigger PipelineTrigger
}

// NewWorkflowService 构造：显式传 db/trigger 供单测注入，生产传 nil
func NewWorkflowService(db *gorm.DB, trigger PipelineTrigger) *WorkflowService {
	return &WorkflowService{db: db, trigger: trigger}
}

// dbh 当前会话：显式注入优先，否则底座全局库
func (s *WorkflowService) dbh() *gorm.DB {
	if s.db != nil {
		return s.db
	}
	return global.GVA_DB
}

// ValidateDefinition 定义校验（纯逻辑，可单测）：JSON 可解析、状态引用闭合、start 在状态集
func ValidateDefinition(statesJSON, transitionsJSON, startState string) ([]model.WfState, []model.WfTransition, error) {
	var states []model.WfState
	if err := json.Unmarshal([]byte(statesJSON), &states); err != nil {
		return nil, nil, errx.New(ErrCodeWfStructInvalid, "states JSON 不合法")
	}
	var transitions []model.WfTransition
	if err := json.Unmarshal([]byte(transitionsJSON), &transitions); err != nil {
		return nil, nil, errx.New(ErrCodeWfStructInvalid, "transitions JSON 不合法")
	}
	if len(states) == 0 {
		return nil, nil, errx.New(ErrCodeWfStructInvalid, "至少一个状态节点")
	}
	names := map[string]bool{}
	for _, s := range states {
		if strings.TrimSpace(s.Name) == "" {
			return nil, nil, errx.New(ErrCodeWfStructInvalid, "状态节点缺少名称")
		}
		if names[s.Name] {
			return nil, nil, errx.New(ErrCodeWfStructInvalid, "状态重名: "+s.Name)
		}
		names[s.Name] = true
	}
	if !names[startState] {
		return nil, nil, errx.New(ErrCodeWfStructInvalid, "起始状态不在状态集: "+startState)
	}
	for _, t := range transitions {
		if !names[t.From] || !names[t.To] {
			return nil, nil, errx.New(ErrCodeWfStructInvalid,
				fmt.Sprintf("迁移 %s --%s--> %s 引用了不存在的状态", t.From, t.Action, t.To))
		}
	}
	return states, transitions, nil
}

// NextState 迁移推导（纯逻辑，可单测）：from+action 命中的目标状态
func NextState(transitions []model.WfTransition, from, action string) (string, bool) {
	for _, t := range transitions {
		if t.From == from && t.Action == action {
			return t.To, true
		}
	}
	return "", false
}

// CreateDefinition 创建定义
func (s *WorkflowService) CreateDefinition(d *model.WfDefinition) error {
	if strings.TrimSpace(d.Name) == "" {
		return errx.New(ErrCodeWfNameRequired, "定义名称不能为空")
	}
	if _, _, err := ValidateDefinition(d.States, d.Transitions, d.StartState); err != nil {
		return err
	}
	count, err := s.repo.DefinitionCountByName(s.dbh(), d.Name)
	if err != nil {
		return errx.Wrap(ErrCodeWfStore, "查询定义失败", err)
	}
	if count > 0 {
		return errx.New(ErrCodeWfNameRequired, "定义已存在: "+d.Name)
	}
	if err := s.repo.CreateDefinition(s.dbh(), d); err != nil {
		return errx.Wrap(ErrCodeWfStore, "创建定义失败", err)
	}
	return nil
}

// UpdateDefinition 更新定义。SubmitAction 每次动作实时读现行定义（并非快照），改定义会
// 改变在途工单的迁移与审批约束——与 DeleteDefinition 同口径：存在进行中实例时拒绝
// （原注释声称"进行中实例不受影响"与实现不符）
func (s *WorkflowService) UpdateDefinition(d *model.WfDefinition) error {
	if d == nil || d.ID == 0 || strings.TrimSpace(d.Name) == "" {
		return errx.New(ErrCodeWfNameRequired, "定义ID与名称不能为空")
	}
	if _, _, err := ValidateDefinition(d.States, d.Transitions, d.StartState); err != nil {
		return err
	}
	running, err := s.repo.RunningInstanceCount(s.dbh(), d.ID)
	if err != nil {
		return errx.Wrap(ErrCodeWfStore, "查询进行中工单失败", err)
	}
	if running > 0 {
		return errx.New(ErrCodeWfStructInvalid, fmt.Sprintf("仍有 %d 个进行中工单使用该定义，处理完毕后才能更新", running))
	}
	if err := s.repo.UpdateDefinitionColumns(s.dbh(), d.ID, map[string]any{
		"name": d.Name, "states": d.States, "transitions": d.Transitions,
		"start_state": d.StartState, "notes": d.Notes,
	}); err != nil {
		return errx.Wrap(ErrCodeWfStore, "更新定义失败", err)
	}
	return nil
}

// ToggleDefinition 启用/停用定义（仅改开关位，不动结构，无进行中实例限制）
func (s *WorkflowService) ToggleDefinition(id uint, enabled bool) error {
	affected, err := s.repo.UpdateDefinitionEnabled(s.dbh(), id, enabled)
	if err != nil {
		return errx.Wrap(ErrCodeWfStore, "更新定义失败", err)
	}
	if affected == 0 {
		return errx.New(ErrCodeWfNotFound, "工单定义不存在")
	}
	return nil
}

// DeleteDefinition 删除定义（有进行中实例时拒绝）
func (s *WorkflowService) DeleteDefinition(id uint) error {
	running, err := s.repo.RunningInstanceCount(s.dbh(), id)
	if err != nil {
		return errx.Wrap(ErrCodeWfStore, "查询进行中工单失败", err)
	}
	if running > 0 {
		return errx.New(ErrCodeWfStructInvalid, fmt.Sprintf("仍有 %d 个进行中工单，不可删除", running))
	}
	if err := s.repo.DeleteDefinition(s.dbh(), id); err != nil {
		return errx.Wrap(ErrCodeWfStore, "删除定义失败", err)
	}
	return nil
}

// GetDefinitionList 定义列表
func (s *WorkflowService) GetDefinitionList(keyword string) ([]*model.WfDefinition, error) {
	list, err := s.repo.DefinitionList(s.dbh(), keyword)
	if err != nil {
		return nil, errx.Wrap(ErrCodeWfStore, "查询定义列表失败", err)
	}
	return list, nil
}

// StartInstance 发起工单：校验定义 → 建实例 → 初始流转记录
func (s *WorkflowService) StartInstance(definitionID uint, title, bizType, params, operator string, userID uint) (*model.WfInstance, error) {
	d, err := s.repo.FindDefinition(s.dbh(), definitionID)
	if err != nil {
		return nil, errx.New(ErrCodeWfNotFound, "工单定义不存在")
	}
	if !d.Enabled {
		return nil, errx.New(ErrCodeWfStructInvalid, "工单定义已停用，无法发起新工单")
	}
	if strings.TrimSpace(title) == "" {
		return nil, errx.New(ErrCodeWfNameRequired, "工单标题不能为空")
	}
	ins := &model.WfInstance{
		DefinitionID: d.ID, DefinitionNm: d.Name,
		Title: title, BizType: bizType, Params: params,
		State: d.StartState, Status: model.InstanceRunning,
		Creator: operator, UserID: userID,
	}
	err = s.dbh().Transaction(func(tx *gorm.DB) error {
		if err := s.repo.CreateInstance(tx, ins); err != nil {
			return errx.Wrap(ErrCodeWfStore, "创建工单失败", err)
		}
		return s.repo.CreateActionLog(tx, &model.WfActionLog{
			InstanceID: ins.ID, FromState: "", Action: "start",
			ToState: d.StartState, Operator: operator,
		})
	})
	if err != nil {
		return nil, err
	}
	return ins, nil
}

// SubmitAction 流转：按当前状态+动作推导；审批节点动作受限（approve/reject）；
// 到达终态自动收档（完成/驳回）；无迁移的动作报错。
func (s *WorkflowService) SubmitAction(instanceID uint, action, comment, operator string) (*model.WfInstance, error) {
	ins, err := s.repo.FindInstance(s.dbh(), instanceID)
	if err != nil {
		return nil, errx.New(ErrCodeWfNotFound, "工单不存在")
	}
	if ins.Status != model.InstanceRunning {
		return nil, errx.New(ErrCodeWfActionInvalid, "工单已结束，不可操作")
	}
	if action == "cancel" {
		now := time.Now()
		err = s.dbh().Transaction(func(tx *gorm.DB) error {
			// 乐观锁：仅 running 状态可撤回（并发流转/收档竞态防护）
			affected, err := s.repo.CancelIfRunning(tx, instanceID, now)
			if err != nil {
				return err
			}
			if affected == 0 {
				return errx.New(ErrCodeWfActionInvalid, "工单状态已变化，请刷新后重试")
			}
			return s.repo.CreateActionLog(tx, &model.WfActionLog{
				InstanceID: instanceID, FromState: ins.State, Action: "cancel",
				ToState: ins.State, Operator: operator, Comment: comment,
			})
		})
		if err != nil {
			return nil, err
		}
		return ins, nil
	}
	d, err := s.repo.FindDefinition(s.dbh(), ins.DefinitionID)
	if err != nil {
		return nil, errx.New(ErrCodeWfNotFound, "工单定义已被删除，无法流转")
	}
	states, transitions, err := ValidateDefinition(d.States, d.Transitions, d.StartState)
	if err != nil {
		return nil, err
	}
	// 当前节点审批约束
	for _, st := range states {
		if st.Name == ins.State && st.IsApproval {
			if action != "approve" && action != "reject" {
				return nil, errx.New(ErrCodeWfActionInvalid, "审批节点仅允许 approve/reject")
			}
		}
	}
	to, ok := NextState(transitions, ins.State, action)
	if !ok {
		return nil, errx.New(ErrCodeWfActionInvalid,
			fmt.Sprintf("状态「%s」不接受动作「%s」", ins.State, action))
	}
	// 目标终态判定
	toFinal := false
	toRejected := false
	for _, st := range states {
		if st.Name == to && st.IsFinal {
			toFinal = true
			if action == "reject" {
				toRejected = true
			}
		}
	}
	updates := map[string]any{"state": to}
	finishedNow := false
	if toFinal {
		now := time.Now()
		if toRejected {
			updates["status"] = model.InstanceRejected
		} else {
			updates["status"] = model.InstanceFinished
			finishedNow = true
		}
		updates["finished_at"] = &now
	}
	// 乐观锁：Where 携带读取时的 state/status，并发流转（两人同时审批/审批与撤回竞态）
	// 只有一个提交生效，其余报"状态已变化"——同时防发版钩子重复触发
	// （条件更新与流转日志须同事务：日志失败则流转一并回滚）
	err = s.dbh().Transaction(func(tx *gorm.DB) error {
		affected, err := s.repo.TransitionIfState(tx, instanceID, ins.State, updates)
		if err != nil {
			return err
		}
		if affected == 0 {
			return errx.New(ErrCodeWfActionInvalid, "工单状态已变化，请刷新后重试")
		}
		return s.repo.CreateActionLog(tx, &model.WfActionLog{
			InstanceID: instanceID, FromState: ins.State, Action: action,
			ToState: to, Operator: operator, Comment: comment,
		})
	})
	if err != nil {
		if errx.CodeOf(err) != 0 {
			return nil, err
		}
		return nil, errx.Wrap(ErrCodeWfStore, "工单流转失败", err)
	}
	// 发版闭环：审批完成（非驳回）的 release 工单自动触发流水线（M3 末项）
	if finishedNow && ins.BizType == BizTypeRelease {
		s.fireReleaseHook(ins, operator)
	}
	return ins, nil
}

// BizTypeRelease 发版工单业务类型约定
const BizTypeRelease = "release"

// ReleaseHookResult 发版钩子结果（测试与日志共用）
type ReleaseHookResult struct {
	Triggered bool
	BuildNo   int
	Err       error
}

// fireReleaseHook 解析工单参数触发流水线构建，结果落流转记录
// params 约定：{"pipelineId":1,"params":{"k":"v"}}
func (s *WorkflowService) fireReleaseHook(ins *model.WfInstance, operator string) {
	var p struct {
		PipelineID uint              `json:"pipelineId"`
		Params     map[string]string `json:"params"`
	}
	if err := json.Unmarshal([]byte(ins.Params), &p); err != nil || p.PipelineID == 0 {
		s.appendHookLog(ins.ID, "发版钩子未触发：参数缺少 pipelineId")
		return
	}
	// 仅经过审批节点的工单才触发发版：原只看"非驳回到达终态"，自定义定义中
	// abort/跳过等任意动作直达终态同样会触发流水线构建（误发生产）
	approved, err := s.repo.ApproveLogCount(s.dbh(), ins.ID)
	if err != nil || approved == 0 {
		s.appendHookLog(ins.ID, "发版钩子未触发：未经过审批（无 approve 流转记录）")
		return
	}
	build, err := s.triggerFunc().CreateBuild(p.PipelineID, p.Params, "工单#"+fmt.Sprintf("%d", ins.ID), ins.UserID, "")
	if err != nil {
		s.appendHookLog(ins.ID, "发版钩子触发失败: "+err.Error())
		return
	}
	s.appendHookLog(ins.ID, fmt.Sprintf("发版钩子已触发流水线「%s」构建 #%d", build.PipelineName, build.BuildNo))
}

// triggerFunc 触发器解析：显式注入优先，否则默认流水线触发器
func (s *WorkflowService) triggerFunc() PipelineTrigger {
	if s.trigger != nil {
		return s.trigger
	}
	return defaultTrigger()
}

// appendHookLog 钩子结果记录（action=hook）
func (s *WorkflowService) appendHookLog(instanceID uint, msg string) {
	_ = s.repo.CreateActionLog(s.dbh(), &model.WfActionLog{
		InstanceID: instanceID, Action: "hook",
		ToState: "pipeline", Operator: "system", Comment: msg,
	})
}

// GetInstanceList 工单分页（creator/status 过滤；普通用户仅自己发起的，888 全量）
func (s *WorkflowService) GetInstanceList(page, pageSize int, creator, status string, userID uint, isSuperAdmin bool) ([]*model.WfInstance, int64, error) {
	// 既有口径：缺省页 10（非 data.Page 缺省 20），钳位 [1,100]，page 归一 ≥1
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	list, total, err := s.repo.InstanceList(s.dbh(), page, pageSize, creator, status, userID, !isSuperAdmin)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeWfStore, "查询工单失败", err)
	}
	return list, total, nil
}

// GetInstanceLogs 流转记录
// 工单流转记录（属主校验：888 放行；其他用户仅可读自己的工单——
// 原 无归属校验且 GET 路由对 9528 开放，遍历 id 即可读他人工单的审批/评论/hook 日志）
func (s *WorkflowService) GetInstanceLogs(instanceID uint, userID uint, isSuperAdmin bool) ([]*model.WfActionLog, error) {
	if !isSuperAdmin {
		count, err := s.repo.InstanceOwnerCount(s.dbh(), instanceID, userID)
		if err != nil {
			return nil, errx.Wrap(ErrCodeWfStore, "查询工单归属失败", err)
		}
		if count == 0 {
			return nil, errx.New(ErrCodeWfNotFound, "工单不存在或无权访问")
		}
	}
	list, err := s.repo.ActionLogList(s.dbh(), instanceID)
	if err != nil {
		return nil, errx.Wrap(ErrCodeWfStore, "查询流转记录失败", err)
	}
	return list, nil
}
