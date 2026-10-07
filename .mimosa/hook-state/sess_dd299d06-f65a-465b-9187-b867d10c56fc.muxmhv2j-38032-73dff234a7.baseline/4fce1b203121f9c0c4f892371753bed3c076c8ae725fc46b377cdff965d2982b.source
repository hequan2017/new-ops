// Package model 白泽工单引擎数据模型（M7：Workflow→State→Transition，husky 蓝本）
// 定义用 states/transitions JSON 灵活建模；实例对定义做快照引用（改定义不影响进行中工单）。
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// 实例状态
const (
	InstanceRunning  = "进行中"
	InstanceFinished = "已完成"
	InstanceRejected = "已驳回"
	InstanceCanceled = "已取消"
)

// WfState 状态节点（isApproval=true 需人工审批：动作仅 approve/reject）
type WfState struct {
	Name       string `json:"name"`
	IsApproval bool   `json:"isApproval"`
	IsFinal    bool   `json:"isFinal"` // 终态（到达即收档完成）
}

// WfTransition 迁移（from + action → to）
type WfTransition struct {
	From   string `json:"from"`
	Action string `json:"action"`
	To     string `json:"to"`
}

// WfDefinition 工单定义
type WfDefinition struct {
	global.GVA_MODEL
	Name        string `json:"name" gorm:"comment:定义名称;unique" binding:"required"`
	States      string `json:"states" gorm:"type:text;comment:状态节点JSON"`
	Transitions string `json:"transitions" gorm:"type:text;comment:迁移JSON"`
	StartState  string `json:"startState" gorm:"comment:起始状态"`
	Notes       string `json:"notes" gorm:"type:text;comment:备注"`
}

// WfInstance 工单实例
type WfInstance struct {
	global.GVA_MODEL
	DefinitionID uint       `json:"definitionId" gorm:"comment:定义ID"`
	DefinitionNm string     `json:"definitionName" gorm:"comment:定义名快照"`
	Title        string     `json:"title" gorm:"comment:工单标题"`
	BizType      string     `json:"bizType" gorm:"comment:业务类型(如 release/sql)"`
	Params       string     `json:"params" gorm:"type:text;comment:业务参数JSON"`
	State        string     `json:"state" gorm:"comment:当前状态"`
	Status       string     `json:"status" gorm:"comment:实例状态;default:进行中;index"`
	Creator      string     `json:"creator" gorm:"comment:发起人"`
	UserID       uint       `json:"userId" gorm:"comment:发起人ID;index"`
	CreatedAt    time.Time  `json:"createdAt"`
	FinishedAt   *time.Time `json:"finishedAt" gorm:"comment:结束时间"`
}

// WfActionLog 流转记录（每步动作与意见）
type WfActionLog struct {
	global.GVA_MODEL
	InstanceID uint      `json:"instanceId" gorm:"comment:工单ID;index"`
	FromState  string    `json:"fromState" gorm:"comment:原状态"`
	Action     string    `json:"action" gorm:"comment:动作"`
	ToState    string    `json:"toState" gorm:"comment:新状态"`
	Operator   string    `json:"operator" gorm:"comment:操作人"`
	Comment    string    `json:"comment" gorm:"type:text;comment:意见"`
	CreatedAt  time.Time `json:"createdAt"`
}
