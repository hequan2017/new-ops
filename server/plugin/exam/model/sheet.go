// Package model exam：答卷
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// ExamAnswerSheet 答卷（券码登录创建；答案幂等保存 + 乐观锁版本）
type ExamAnswerSheet struct {
	global.GVA_MODEL
	ExamId      uint       `json:"examId" gorm:"index;comment:考试ID"`
	CandidateId uint       `json:"candidateId" gorm:"index;comment:考生ID"`
	AnswersJSON string     `json:"answersJSON" gorm:"type:mediumtext;comment:答案{questionId:answer}"`
	Version     int        `json:"version" gorm:"default:0;comment:乐观锁版本"`
	StartedAt   time.Time  `json:"startedAt" gorm:"comment:开始作答时间"`
	SubmittedAt *time.Time `json:"submittedAt" gorm:"comment:交卷时间"`
	Status      string     `json:"status" gorm:"type:varchar(16);default:ongoing;index;comment:ongoing/submitted"`
}

func (ExamAnswerSheet) TableName() string { return "exam_answer_sheet" }
