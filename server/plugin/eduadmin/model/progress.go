// Package model eduadmin 教育管理一模通管：工单进度与督办（需求 3.6-36）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduTicketTimeline 工单时间线（办理单位/人员/时间/进展）
type EduTicketTimeline struct {
	global.GVA_MODEL
	TicketId   uint   `json:"ticketId" gorm:"index;comment:工单ID"`
	Dept       string `json:"dept" gorm:"type:varchar(100);comment:办理单位"`
	Operator   string `json:"operator" gorm:"type:varchar(100);comment:办理人员"`
	Content    string `json:"content" gorm:"type:varchar(500);comment:进展说明"`
	OccurredAt string `json:"occurredAt" gorm:"type:varchar(32);comment:发生时间"`
}

func (EduTicketTimeline) TableName() string { return "edu_ticket_timeline" }

// EduTicketUrge 督办记录（临近时限提醒/超时自动督办）
type EduTicketUrge struct {
	global.GVA_MODEL
	TicketId  uint   `json:"ticketId" gorm:"uniqueIndex:idx_urge_ticket_level;comment:工单ID"`
	Level     string `json:"level" gorm:"type:varchar(16);uniqueIndex:idx_urge_ticket_level;comment:临近/超时"`
	DeadlineAt string `json:"deadlineAt" gorm:"type:varchar(32);comment:时限时点"`
	UrgedAt   string `json:"urgedAt" gorm:"type:varchar(32);comment:督办时间"`
}

func (EduTicketUrge) TableName() string { return "edu_ticket_urge" }
