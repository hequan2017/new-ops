// Package service 白泽工单引擎（M7）：定义校验 + 实例流转
// workflow 插件错误码段：1900-1999（DEV_PLAN 3.6）
package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/workflow/model"
	"gorm.io/gorm"
)

// 错误码
const (
	ErrCodeWfNameRequired   = 1901
	ErrCodeWfStructInvalid  = 1902
	ErrCodeWfNotFound       = 1903
	ErrCodeWfActionInvalid  = 1904
)

func newWfErr(code int, msg string) error { return &wfError{Code: code, Msg: msg} }

type wfError struct {
	Code int
	Msg  string
}

func (e *wfError) Error() string { return e.Msg }

// WorkflowService 工单服务
type WorkflowService struct{}

// ValidateDefinition 定义校验（纯逻辑，可单测）：JSON 可解析、状态引用闭合、start 在状态集
func ValidateDefinition(statesJSON, transitionsJSON, startState string) ([]model.WfState, []model.WfTransition, error) {
	var states []model.WfState
	if err := json.Unmarshal([]byte(statesJSON), &states); err != nil {
		return nil, nil, newWfErr(ErrCodeWfStructInvalid, "states JSON 不合法")
	}
	var transitions []model.WfTransition
	if err := json.Unmarshal([]byte(transitionsJSON), &transitions); err != nil {
		return nil, nil, newWfErr(ErrCodeWfStructInvalid, "transitions JSON 不合法")
	}
	if len(states) == 0 {
		return nil, nil, newWfErr(ErrCodeWfStructInvalid, "至少一个状态节点")
	}
	names := map[string]bool{}
	for _, s := range states {
		if strings.TrimSpace(s.Name) == "" {
			return nil, nil, newWfErr(ErrCodeWfStructInvalid, "状态节点缺少名称")
		}
		if names[s.Name] {
			return nil, nil, newWfErr(ErrCodeWfStructInvalid, "状态重名: "+s.Name)
		}
		names[s.Name] = true
	}
	if !names[startState] {
		return nil, nil, newWfErr(ErrCodeWfStructInvalid, "起始状态不在状态集: "+startState)
	}
	for _, t := range transitions {
		if !names[t.From] || !names[t.To] {
			return nil, nil, newWfErr(ErrCodeWfStructInvalid,
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
		return newWfErr(ErrCodeWfNameRequired, "定义名称不能为空")
	}
	if _, _, err := ValidateDefinition(d.States, d.Transitions, d.StartState); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.WfDefinition{}).Where("name = ?", d.Name).Count(&count)
	if count > 0 {
		return newWfErr(ErrCodeWfNameRequired, "定义已存在: "+d.Name)
	}
	return global.GVA_DB.Create(d).Error
}

// UpdateDefinition 更新定义（进行中实例不受影响——实例只按定义ID读历史快照语义）
func (s *WorkflowService) UpdateDefinition(d *model.WfDefinition) error {
	if d == nil || d.ID == 0 || strings.TrimSpace(d.Name) == "" {
		return newWfErr(ErrCodeWfNameRequired, "定义ID与名称不能为空")
	}
	if _, _, err := ValidateDefinition(d.States, d.Transitions, d.StartState); err != nil {
		return err
	}
	return global.GVA_DB.Model(&model.WfDefinition{}).Where("id = ?", d.ID).Updates(map[string]any{
		"name": d.Name, "states": d.States, "transitions": d.Transitions,
		"start_state": d.StartState, "notes": d.Notes,
	}).Error
}

// DeleteDefinition 删除定义（有进行中实例时拒绝）
func (s *WorkflowService) DeleteDefinition(id uint) error {
	var running int64
	global.GVA_DB.Model(&model.WfInstance{}).
		Where("definition_id = ? AND status = ?", id, model.InstanceRunning).Count(&running)
	if running > 0 {
		return newWfErr(ErrCodeWfStructInvalid, fmt.Sprintf("仍有 %d 个进行中工单，不可删除", running))
	}
	return global.GVA_DB.Delete(&model.WfDefinition{}, id).Error
}

// GetDefinitionList 定义列表
func (s *WorkflowService) GetDefinitionList(keyword string) (list []*model.WfDefinition, err error) {
	db := global.GVA_DB.Model(&model.WfDefinition{})
	if keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	err = db.Order("id DESC").Limit(200).Find(&list).Error
	return list, err
}

// StartInstance 发起工单：校验定义 → 建实例 → 初始流转记录
func (s *WorkflowService) StartInstance(definitionID uint, title, bizType, params, operator string, userID uint) (*model.WfInstance, error) {
	var d model.WfDefinition
	if err := global.GVA_DB.First(&d, definitionID).Error; err != nil {
		return nil, newWfErr(ErrCodeWfNotFound, "工单定义不存在")
	}
	if strings.TrimSpace(title) == "" {
		return nil, newWfErr(ErrCodeWfNameRequired, "工单标题不能为空")
	}
	ins := &model.WfInstance{
		DefinitionID: d.ID, DefinitionNm: d.Name,
		Title: title, BizType: bizType, Params: params,
		State: d.StartState, Status: model.InstanceRunning,
		Creator: operator, UserID: userID,
	}
	return ins, global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ins).Error; err != nil {
			return err
		}
		return tx.Create(&model.WfActionLog{
			InstanceID: ins.ID, FromState: "", Action: "start",
			ToState: d.StartState, Operator: operator,
		}).Error
	})
}

// SubmitAction 流转：按当前状态+动作推导；审批节点动作受限（approve/reject）；
// 到达终态自动收档（完成/驳回）；无迁移的动作报错。
func (s *WorkflowService) SubmitAction(instanceID uint, action, comment, operator string) (*model.WfInstance, error) {
	var ins model.WfInstance
	if err := global.GVA_DB.First(&ins, instanceID).Error; err != nil {
		return nil, newWfErr(ErrCodeWfNotFound, "工单不存在")
	}
	if ins.Status != model.InstanceRunning {
		return nil, newWfErr(ErrCodeWfActionInvalid, "工单已结束，不可操作")
	}
	if action == "cancel" {
		now := time.Now()
		return &ins, global.GVA_DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&model.WfInstance{}).Where("id = ?", instanceID).Updates(map[string]any{
				"status": model.InstanceCanceled, "finished_at": &now,
			}).Error; err != nil {
				return err
			}
			return tx.Create(&model.WfActionLog{
				InstanceID: instanceID, FromState: ins.State, Action: "cancel",
				ToState: ins.State, Operator: operator, Comment: comment,
			}).Error
		})
	}
	var d model.WfDefinition
	if err := global.GVA_DB.First(&d, ins.DefinitionID).Error; err != nil {
		return nil, newWfErr(ErrCodeWfNotFound, "工单定义已被删除，无法流转")
	}
	states, transitions, err := ValidateDefinition(d.States, d.Transitions, d.StartState)
	if err != nil {
		return nil, err
	}
	// 当前节点审批约束
	for _, st := range states {
		if st.Name == ins.State && st.IsApproval {
			if action != "approve" && action != "reject" {
				return nil, newWfErr(ErrCodeWfActionInvalid, "审批节点仅允许 approve/reject")
			}
		}
	}
	to, ok := NextState(transitions, ins.State, action)
	if !ok {
		return nil, newWfErr(ErrCodeWfActionInvalid,
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
	if toFinal {
		now := time.Now()
		if toRejected {
			updates["status"] = model.InstanceRejected
		} else {
			updates["status"] = model.InstanceFinished
		}
		updates["finished_at"] = &now
	}
	return &ins, global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.WfInstance{}).Where("id = ?", instanceID).Updates(updates).Error; err != nil {
			return err
		}
		return tx.Create(&model.WfActionLog{
			InstanceID: instanceID, FromState: ins.State, Action: action,
			ToState: to, Operator: operator, Comment: comment,
		}).Error
	})
}

// GetInstanceList 工单分页（creator/status 过滤；普通用户仅自己发起的，888 全量）
func (s *WorkflowService) GetInstanceList(page, pageSize int, creator, status string, userID uint, isSuperAdmin bool) (list []*model.WfInstance, total int64, err error) {
	db := global.GVA_DB.Model(&model.WfInstance{})
	if !isSuperAdmin {
		db = db.Where("user_id = ?", userID)
	}
	if creator != "" {
		db = db.Where("creator LIKE ?", "%"+creator+"%")
	}
	if status != "" {
		db = db.Where("status = ?", status)
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

// GetInstanceLogs 流转记录
func (s *WorkflowService) GetInstanceLogs(instanceID uint) (list []*model.WfActionLog, err error) {
	err = global.GVA_DB.Where("instance_id = ?", instanceID).Order("id ASC").Limit(100).Find(&list).Error
	return list, err
}
