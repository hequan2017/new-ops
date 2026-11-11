// Package model crossauth 跨校访问认证与安全：三级审计（需求 3.7-14）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CaAuditEvent 统一审计事件（校级/院系级/用户级 分层过滤）
type CaAuditEvent struct {
	global.GVA_MODEL
	Level      string `json:"level" gorm:"type:varchar(16);index;comment:校级/院系级/用户级"`
	School     string `json:"school" gorm:"type:varchar(100);index;comment:院校"`
	Dept       string `json:"dept" gorm:"type:varchar(100);index;comment:院系"`
	UserName   string `json:"userName" gorm:"type:varchar(100);index;comment:用户"`
	Action     string `json:"action" gorm:"type:varchar(100);comment:动作(登录/授权/访问/导出等)"`
	Detail     string `json:"detail" gorm:"type:varchar(500);comment:详情"`
	OccurredAt string `json:"occurredAt" gorm:"type:varchar(32);comment:发生时间(CST)"`
}

func (CaAuditEvent) TableName() string { return "ca_audit_event" }
