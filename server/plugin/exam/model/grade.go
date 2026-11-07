// Package model exam：判分明细
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ExamGrade 判分明细（客观 auto / 主观实操 llm）
type ExamGrade struct {
	global.GVA_MODEL
	SheetId    uint    `json:"sheetId" gorm:"index;comment:答卷ID"`
	QuestionId uint    `json:"questionId" gorm:"index;comment:题目ID"`
	Type       string  `json:"type" gorm:"type:varchar(8);comment:题型"`
	Answer     string  `json:"answer" gorm:"type:varchar(1000);comment:考生答案"`
	Standard   string  `json:"standard" gorm:"type:varchar(1000);comment:标准答案"`
	Score      float64 `json:"score" gorm:"comment:得分"`
	FullScore  float64 `json:"fullScore" gorm:"comment:满分"`
	Method     string  `json:"method" gorm:"type:varchar(8);comment:auto/llm"`
	Detail     string  `json:"detail" gorm:"type:varchar(500);comment:判分说明"`
}

func (ExamGrade) TableName() string { return "exam_grade" }
