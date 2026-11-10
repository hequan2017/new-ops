// Package model eduadmin 教育管理一模通管：人才预测（需求 3.6-14）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduTalentStat 年度人才统计（招聘/离职/在职）
type EduTalentStat struct {
	global.GVA_MODEL
	Year       int `json:"year" gorm:"uniqueIndex;comment:年度"`
	NewHires   int `json:"newHires" gorm:"comment:新聘人数"`
	Departures int `json:"departures" gorm:"comment:离职人数"`
	Headcount  int `json:"headcount" gorm:"comment:期末在职人数"`
}

func (EduTalentStat) TableName() string { return "edu_talent_stat" }
