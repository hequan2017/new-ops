// Package model eduadmin 教育管理一模通管：反馈闭环（需求 3.6-36）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduTicketReview 工单满意度评价（好差评）
type EduTicketReview struct {
	global.GVA_MODEL
	TicketId  uint   `json:"ticketId" gorm:"uniqueIndex;comment:工单ID"`
	Score     string `json:"score" gorm:"type:varchar(16);index;comment:好评/差评"`
	Comment   string `json:"comment" gorm:"type:varchar(500);comment:评价内容"`
	ReviewedAt string `json:"reviewedAt" gorm:"type:varchar(32);comment:评价时间"`
}

func (EduTicketReview) TableName() string { return "edu_ticket_review" }
