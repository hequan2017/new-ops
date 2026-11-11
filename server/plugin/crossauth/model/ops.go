// Package model crossauth 跨校访问认证与安全：跨校联合运维（需求 3.7-15）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CaOpsTicket 运维工单（状态机：提交→分配→处理→闭环）
type CaOpsTicket struct {
	global.GVA_MODEL
	TicketNo string `json:"ticketNo" gorm:"type:varchar(64);uniqueIndex;comment:工单号"`
	School   string `json:"school" gorm:"type:varchar(100);index;comment:报障院校"`
	Title    string `json:"title" gorm:"type:varchar(300);comment:故障标题"`
	Status   string `json:"status" gorm:"type:varchar(16);index;comment:提交/分配/处理/闭环"`
	Assignee string `json:"assignee" gorm:"type:varchar(100);comment:承接运维方"`
}

func (CaOpsTicket) TableName() string { return "ca_ops_ticket" }

// CaOpsMessage 协同消息（文字口径）
type CaOpsMessage struct {
	global.GVA_MODEL
	TicketId uint   `json:"ticketId" gorm:"index;comment:工单ID"`
	From     string `json:"from" gorm:"type:varchar(100);comment:发送方(院校/运维)"`
	Content  string `json:"content" gorm:"type:varchar(1000);comment:消息内容(文字口径)"`
	SentAt   string `json:"sentAt" gorm:"type:varchar(32);comment:发送时间(CST)"`
}

func (CaOpsMessage) TableName() string { return "ca_ops_message" }

// CaRemoteAssist 远程协助授权记录
type CaRemoteAssist struct {
	global.GVA_MODEL
	TicketId    uint   `json:"ticketId" gorm:"index;comment:工单ID"`
	Requester   string `json:"requester" gorm:"type:varchar(100);comment:申请运维方"`
	Approver    string `json:"approver" gorm:"type:varchar(100);comment:授权院校方"`
	Scope       string `json:"scope" gorm:"type:varchar(300);comment:授权范围(只读/操作级)"`
	AuthorizedAt string `json:"authorizedAt" gorm:"type:varchar(32);comment:授权时间(CST)"`
	Revoked     bool   `json:"revoked" gorm:"index;comment:已撤销"`
}

func (CaRemoteAssist) TableName() string { return "ca_remote_assist" }
