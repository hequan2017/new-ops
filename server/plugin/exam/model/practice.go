// Package model exam：实操执行留痕
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ExamPracticeLog 实操执行留痕（采购件④ 过程性判分依据）
type ExamPracticeLog struct {
	global.GVA_MODEL
	SheetId    uint   `json:"sheetId" gorm:"index;comment:答卷ID"`
	QuestionId uint   `json:"questionId" gorm:"index;comment:实操题ID"`
	Channel    string `json:"channel" gorm:"type:varchar(16);comment:operator/agent"`
	Target     string `json:"target" gorm:"type:varchar(128);comment:算子名/智能体模型目标"`
	Input      string `json:"input" gorm:"type:text;comment:输入"`
	Output     string `json:"output" gorm:"type:mediumtext;comment:输出"`
	LatencyMs  int64  `json:"latencyMs" gorm:"comment:耗时毫秒"`
	Ok         bool   `json:"ok" gorm:"index;comment:是否成功"`
	ErrMsg     string `json:"errMsg" gorm:"type:varchar(500);comment:错误"`
}

func (ExamPracticeLog) TableName() string { return "exam_practice_log" }
