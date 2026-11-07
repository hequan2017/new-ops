// Package model exam：题库
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ExamQuestion 考题（采购件③ 五题型：single 单选/multiple 多选/judge 判断/short 简答/practice 实操）
type ExamQuestion struct {
	global.GVA_MODEL
	Type           string  `json:"type" gorm:"type:varchar(8);index;comment:single/multiple/judge/short/practice"`
	Stem           string  `json:"stem" gorm:"type:text;comment:题干"`
	OptionsJSON    string  `json:"optionsJSON" gorm:"type:text;comment:选项JSON数组(选择类)"`
	Answer         string  `json:"answer" gorm:"type:varchar(1000);comment:标准答案"`
	Score          float64 `json:"score" gorm:"default:1;comment:分值"`
	KnowledgePoint string  `json:"knowledgePoint" gorm:"type:varchar(64);index;comment:知识点"`
	Difficulty     int     `json:"difficulty" gorm:"default:1;comment:难度1-5"`
}

func (ExamQuestion) TableName() string { return "exam_question" }
