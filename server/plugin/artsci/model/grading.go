// Package model artsci 学科工具集：作业批改记录
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ArtsciGrading 智能作业批改与反馈记录（需求 3.3-8）
type ArtsciGrading struct {
	global.GVA_MODEL
	Title       string  `json:"title" gorm:"type:varchar(200);comment:作业标题"`
	Rubric      string  `json:"rubric" gorm:"type:text;comment:评分标准/要点"`
	StudentText string  `json:"studentText" gorm:"type:mediumtext;comment:学生作答"`
	Score       float64 `json:"score" gorm:"comment:得分(0-100)"`
	Comment     string  `json:"comment" gorm:"type:varchar(1000);comment:总评"`
	Suggestions string  `json:"suggestions" gorm:"type:varchar(1000);comment:改进建议(JSON 数组)"`
	Engine      string  `json:"engine" gorm:"type:varchar(16);default:rule;index;comment:llm/rule"`
	ModelId     uint    `json:"modelId" gorm:"default:0;comment:llmops 模型ID(0=规则口径)"`
	Ok          bool    `json:"ok" gorm:"index;comment:批改是否成功"`
	ErrMsg      string  `json:"errMsg" gorm:"type:varchar(500);comment:错误"`
}

func (ArtsciGrading) TableName() string { return "artsci_grading" }
