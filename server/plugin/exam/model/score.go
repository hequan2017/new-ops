// Package model exam：成绩与等级规则
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ExamScore 成绩（判分发布后落档；sheetId 唯一）
type ExamScore struct {
	global.GVA_MODEL
	ExamId      uint    `json:"examId" gorm:"index;comment:考试ID"`
	SheetId     uint    `json:"sheetId" gorm:"uniqueIndex;comment:答卷ID"`
	CandidateId uint    `json:"candidateId" gorm:"index;comment:考生ID"`
	Total       float64 `json:"total" gorm:"comment:总分"`
	Full        float64 `json:"full" gorm:"comment:卷面满分"`
	Level       string  `json:"level" gorm:"type:varchar(16);comment:等级"`
}

func (ExamScore) TableName() string { return "exam_score" }

// ExamLevelRule 等级分数线（examId=0 为全局默认；minRatio 得分率 0-1）
type ExamLevelRule struct {
	global.GVA_MODEL
	ExamId   uint    `json:"examId" gorm:"index;comment:考试ID(0=全局默认)"`
	Level    string  `json:"level" gorm:"type:varchar(16);comment:等级名"`
	MinRatio float64 `json:"minRatio" gorm:"comment:最低得分率 0-1"`
}

func (ExamLevelRule) TableName() string { return "exam_level_rule" }
