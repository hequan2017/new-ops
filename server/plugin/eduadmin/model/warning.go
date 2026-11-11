// Package model eduadmin 教育管理一模通管：办学预警（需求 3.6-25）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduRunIndex 办学指标（招生/就业/经费）
type EduRunIndex struct {
	global.GVA_MODEL
	Name        string  `json:"name" gorm:"type:varchar(100);index;comment:指标名(招生完成率/就业率/经费到位率)"`
	Year        int     `json:"year" gorm:"index;comment:年度"`
	Value       float64 `json:"value" gorm:"comment:指标值(0-1 比例)"`
	Threshold   float64 `json:"threshold" gorm:"comment:预警阈值"`
	HigherBetter bool   `json:"higherBetter" gorm:"default:true;comment:是否越高越好"`
}

func (EduRunIndex) TableName() string { return "edu_run_index" }

// EduRunAlert 办学预警记录
type EduRunAlert struct {
	global.GVA_MODEL
	IndexName string  `json:"indexName" gorm:"type:varchar(100);index;comment:指标名"`
	Year      int     `json:"year" gorm:"index;comment:年度"`
	Level     string  `json:"level" gorm:"type:varchar(16);comment:高/中"`
	Value     float64 `json:"value" gorm:"comment:指标值"`
	Threshold float64 `json:"threshold" gorm:"comment:阈值"`
	Advice    string  `json:"advice" gorm:"type:varchar(500);comment:策略建议"`
}

func (EduRunAlert) TableName() string { return "edu_run_alert" }
