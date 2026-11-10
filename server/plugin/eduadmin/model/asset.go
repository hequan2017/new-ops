// Package model eduadmin 教育管理一模通管：资产故障预测（需求 3.6-18）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduAsset 设备资产（传感器指标与阈值）
type EduAsset struct {
	global.GVA_MODEL
	Name     string  `json:"name" gorm:"type:varchar(200);comment:设备名"`
	Location string  `json:"location" gorm:"type:varchar(200);comment:安装位置"`
	Metric   string  `json:"metric" gorm:"type:varchar(32);comment:传感器指标(温度/振动/电流)"`
	Threshold float64 `json:"threshold" gorm:"comment:故障阈值"`
}

func (EduAsset) TableName() string { return "edu_asset" }

// EduSensorReading 传感器时序读数
type EduSensorReading struct {
	global.GVA_MODEL
	AssetId     uint    `json:"assetId" gorm:"index;comment:设备ID"`
	MetricValue float64 `json:"metricValue" gorm:"comment:指标值"`
	RecordedAt  string  `json:"recordedAt" gorm:"type:varchar(32);comment:采集时间"`
}

func (EduSensorReading) TableName() string { return "edu_sensor_reading" }

// EduMaintPlan 维护计划
type EduMaintPlan struct {
	global.GVA_MODEL
	AssetId uint    `json:"assetId" gorm:"index;comment:设备ID"`
	Level   string  `json:"level" gorm:"type:varchar(16);comment:无故障/轻微/严重"`
	Slope   float64 `json:"slope" gorm:"comment:趋势斜率"`
	Latest  float64 `json:"latest" gorm:"comment:最新读数"`
	Plan    string  `json:"plan" gorm:"type:varchar(500);comment:维护计划"`
}

func (EduMaintPlan) TableName() string { return "edu_maint_plan" }
