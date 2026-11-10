// Package model eduadmin 教育管理一模通管：设施维护（需求 3.6-22）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduRepairOrder 维修工单（状态机：待派→处理中→已闭环）
type EduRepairOrder struct {
	global.GVA_MODEL
	OrderNo     string `json:"orderNo" gorm:"type:varchar(64);uniqueIndex;comment:工单号"`
	AssetId     uint   `json:"assetId" gorm:"index;comment:设备ID(0 表示非设备工单)"`
	Description string `json:"description" gorm:"type:varchar(500);comment:故障描述"`
	Level       string `json:"level" gorm:"type:varchar(16);comment:严重/轻微/常规"`
	Team        string `json:"team" gorm:"type:varchar(64);comment:承接班组"`
	Status      string `json:"status" gorm:"type:varchar(16);index;comment:待派/处理中/已闭环"`
}

func (EduRepairOrder) TableName() string { return "edu_repair_order" }
