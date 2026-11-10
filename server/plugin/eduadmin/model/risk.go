// Package model eduadmin 教育管理一模通管：财务风险预警（需求 3.6-17）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduRiskAlert 财务预警记录（支出异常/现金流不足）
type EduRiskAlert struct {
	global.GVA_MODEL
	Type      string  `json:"type" gorm:"type:varchar(16);index;comment:支出异常/现金流不足"`
	RefKey    string  `json:"refKey" gorm:"type:varchar(200);index;comment:关联键(发票号/月份)"`
	Detail    string  `json:"detail" gorm:"type:varchar(500);comment:预警详情"`
	Value     float64 `json:"value" gorm:"comment:触发值"`
	Threshold float64 `json:"threshold" gorm:"comment:阈值(3σ 或现金流下限)"`
	Level     string  `json:"level" gorm:"type:varchar(16);comment:高/中/低"`
	Handled   bool    `json:"handled" gorm:"index;comment:已处理"`
}

func (EduRiskAlert) TableName() string { return "edu_risk_alert" }
