// Package model artsci 学科工具集：自动出题与组卷（需求 3.3-9）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ArtsciQuestion 生成题目
type ArtsciQuestion struct {
	global.GVA_MODEL
	KnowledgePoint string `json:"knowledgePoint" gorm:"type:varchar(100);index;comment:知识点"`
	QType          string `json:"qType" gorm:"type:varchar(16);index;comment:单选/多选/判断/简答"`
	Difficulty     string `json:"difficulty" gorm:"type:varchar(8);default:中;comment:易/中/难"`
	Stem           string `json:"stem" gorm:"type:text;comment:题干"`
	Options        string `json:"options" gorm:"type:text;comment:选项(JSON 数组，客观题)"`
	Answer         string `json:"answer" gorm:"type:varchar(500);comment:参考答案"`
	Score          float64 `json:"score" gorm:"default:0;comment:默认分值"`
	Source         string `json:"source" gorm:"type:varchar(16);default:template;index;comment:template/llm"`
}

func (ArtsciQuestion) TableName() string { return "artsci_question" }

// ArtsciPaper 组卷结果
type ArtsciPaper struct {
	global.GVA_MODEL
	Name        string  `json:"name" gorm:"type:varchar(200);comment:试卷名"`
	QuestionIds string  `json:"questionIds" gorm:"type:text;comment:题目ID CSV"`
	TotalScore  float64 `json:"totalScore" gorm:"comment:总分"`
	RulesJSON   string  `json:"rulesJSON" gorm:"type:text;comment:抽题规则留痕"`
}

func (ArtsciPaper) TableName() string { return "artsci_paper" }
