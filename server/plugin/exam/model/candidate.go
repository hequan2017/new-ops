// Package model exam：考生
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ExamCandidate 考生（券码唯一索引；状态 pending/entered/submitted）
type ExamCandidate struct {
	global.GVA_MODEL
	ExamId     uint   `json:"examId" gorm:"index;comment:考试ID"`
	Name       string `json:"name" gorm:"type:varchar(64);comment:考生姓名"`
	UserKey    string `json:"userKey" gorm:"type:varchar(64);uniqueIndex:uk_exam_user;comment:考生标识(账号/机器码)"`
	TicketCode string `json:"ticketCode" gorm:"type:varchar(32);uniqueIndex;comment:券码"`
	Status     string `json:"status" gorm:"type:varchar(16);default:pending;comment:pending/entered/submitted"`
}

func (ExamCandidate) TableName() string { return "exam_candidate" }
