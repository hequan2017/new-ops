// Package model eduadmin 教育管理一模通管：AI-12345 工单模型（需求 3.6-36）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduTicket AI-12345 工单（状态机：受理→分派→办理→办结→评价→归档）
type EduTicket struct {
	global.GVA_MODEL
	TicketNo       string `json:"ticketNo" gorm:"type:varchar(64);uniqueIndex;comment:唯一编号(T-yyyyMMdd-序号)"`
	Channel        string `json:"channel" gorm:"type:varchar(16);index;comment:网页/企微/小程序"`
	DemandType     string `json:"demandType" gorm:"type:varchar(64);comment:诉求类型(教务/后勤/财务/信息化等)"`
	Content        string `json:"content" gorm:"type:text;comment:诉求内容"`
	AttachmentMeta string `json:"attachmentMeta" gorm:"type:varchar(500);comment:附件元数据(名称/大小/哈希)"`
	DeadlineClass  string `json:"deadlineClass" gorm:"type:varchar(16);comment:简单24h/一般3工作日/复杂7工作日"`
	Status         string `json:"status" gorm:"type:varchar(16);index;comment:受理/分派/办理/办结/评价/归档"`
}

func (EduTicket) TableName() string { return "edu_ticket" }
