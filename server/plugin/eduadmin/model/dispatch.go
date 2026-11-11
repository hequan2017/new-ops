// Package model eduadmin 教育管理一模通管：智能分派（需求 3.6-36）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduDispatchRule 分派规则（关键词→部门，优先级大者优先）
type EduDispatchRule struct {
	global.GVA_MODEL
	Keywords string `json:"keywords" gorm:"type:varchar(500);comment:关键词(逗号分隔)"`
	Dept     string `json:"dept" gorm:"type:varchar(100);index;comment:承办部门"`
	Priority int    `json:"priority" gorm:"default:0;comment:优先级(大者优先)"`
}

func (EduDispatchRule) TableName() string { return "edu_dispatch_rule" }

// EduDispatchLog 分派记录
type EduDispatchLog struct {
	global.GVA_MODEL
	TicketId uint   `json:"ticketId" gorm:"index;comment:工单ID"`
	RuleId   uint   `json:"ruleId" gorm:"comment:命中规则ID(0=人工兜底)"`
	Dept     string `json:"dept" gorm:"type:varchar(100);comment:承办部门"`
	Auto     bool   `json:"auto" gorm:"comment:是否自动分派"`
	Reason   string `json:"reason" gorm:"type:varchar(300);comment:分派说明"`
}

func (EduDispatchLog) TableName() string { return "edu_dispatch_log" }
