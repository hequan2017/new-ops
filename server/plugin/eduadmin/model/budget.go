// Package model eduadmin 教育管理一模通管：智能预算（需求 3.6-15）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduBudget 部门年度预算
type EduBudget struct {
	global.GVA_MODEL
	Year   int     `json:"year" gorm:"index;comment:年度"`
	Dept   string  `json:"dept" gorm:"type:varchar(100);index;comment:部门"`
	Amount float64 `json:"amount" gorm:"comment:预算金额(万元)"`
}

func (EduBudget) TableName() string { return "edu_budget" }
