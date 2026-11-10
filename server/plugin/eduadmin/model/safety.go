// Package model eduadmin 教育管理一模通管：校园安防-环境监测（需求 3.6-23）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduEnvReading 环境监测读数（空气/水质）
type EduEnvReading struct {
	global.GVA_MODEL
	Type       string  `json:"type" gorm:"type:varchar(16);index;comment:空气/水质"`
	Location   string  `json:"location" gorm:"type:varchar(100);comment:监测点"`
	Metric     string  `json:"metric" gorm:"type:varchar(64);comment:指标(PM2.5/CO2/浊度/余氯等)"`
	Value      float64 `json:"value" gorm:"comment:监测值"`
	Threshold  float64 `json:"threshold" gorm:"comment:告警阈值"`
	RecordedAt string  `json:"recordedAt" gorm:"type:varchar(32);comment:采集时间"`
}

func (EduEnvReading) TableName() string { return "edu_env_reading" }

// EduEnvAlert 环境告警记录
type EduEnvAlert struct {
	global.GVA_MODEL
	ReadingId uint    `json:"readingId" gorm:"index;comment:读数ID"`
	Type      string  `json:"type" gorm:"type:varchar(16);comment:空气/水质"`
	Location  string  `json:"location" gorm:"type:varchar(100);comment:监测点"`
	Metric    string  `json:"metric" gorm:"type:varchar(64);comment:指标"`
	Level     string  `json:"level" gorm:"type:varchar(16);comment:高/中"`
	Value     float64 `json:"value" gorm:"comment:监测值"`
	Threshold float64 `json:"threshold" gorm:"comment:阈值"`
	Advice    string  `json:"advice" gorm:"type:varchar(500);comment:健康建议"`
}

func (EduEnvAlert) TableName() string { return "edu_env_alert" }
