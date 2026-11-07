// Package model exam：试卷
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ExamPaper 试卷（采购件③ 组卷：规则抽题后锁定题目与总分）
type ExamPaper struct {
	global.GVA_MODEL
	Name        string  `json:"name" gorm:"comment:试卷名;unique"`
	RuleJSON    string  `json:"ruleJSON" gorm:"type:text;comment:抽题规则JSON"`
	TotalScore  float64 `json:"totalScore" gorm:"comment:总分(组卷回填)"`
	QuestionIds string  `json:"questionIds" gorm:"type:text;comment:锁定题目ID(逗号分隔)"`
	Status      string  `json:"status" gorm:"type:varchar(16);default:draft;comment:draft/ready"`
}

func (ExamPaper) TableName() string { return "exam_paper" }
