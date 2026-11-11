// Package model crossauth 跨校访问认证与安全：应急演练与预案（需求 3.7-16）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CaPlaybook 应急预案（版本管理：更新即版本+1）
type CaPlaybook struct {
	global.GVA_MODEL
	EventName    string `json:"eventName" gorm:"type:varchar(100);uniqueIndex;comment:事件类型"`
	ResponseFlow string `json:"responseFlow" gorm:"type:text;comment:响应流程"`
	Duties       string `json:"duties" gorm:"type:text;comment:责任分工"`
	Measures     string `json:"measures" gorm:"type:text;comment:处置措施"`
	Version      int    `json:"version" gorm:"default:1;comment:版本号"`
	Status       string `json:"status" gorm:"type:varchar(16);comment:启用/停用"`
}

func (CaPlaybook) TableName() string { return "ca_playbook" }

// CaDrillPlan 演练计划与复盘（每年 2 次口径字段）
type CaDrillPlan struct {
	global.GVA_MODEL
	PlaybookName string `json:"playbookName" gorm:"type:varchar(100);index;comment:预案事件"`
	Year         int    `json:"year" gorm:"index;comment:年度"`
	PlannedAt    string `json:"plannedAt" gorm:"type:varchar(32);comment:计划时间"`
	ExecutedAt   string `json:"executedAt" gorm:"type:varchar(32);comment:实际执行时间"`
	Report       string `json:"report" gorm:"type:text;comment:复盘报告"`
	Grade        string `json:"grade" gorm:"type:varchar(16);comment:评定(优/良/需改进)"`
	YearlyTarget int    `json:"yearlyTarget" gorm:"default:2;comment:年度演练次数口径(2 次)"`
}

func (CaDrillPlan) TableName() string { return "ca_drill_plan" }
