// Package model crossauth 跨校访问认证与安全：访问校验与告警（需求 3.7-11）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CaAccessLog 访问日志
type CaAccessLog struct {
	global.GVA_MODEL
	UserName string `json:"userName" gorm:"type:varchar(100);index;comment:用户"`
	School   string `json:"school" gorm:"type:varchar(100);comment:所属校"`
	Resource string `json:"resource" gorm:"type:varchar(200);index;comment:资源"`
	Allowed  bool   `json:"allowed" gorm:"comment:是否放行"`
	Reason   string `json:"reason" gorm:"type:varchar(300);comment:判定原因"`
}

func (CaAccessLog) TableName() string { return "ca_access_log" }

// CaAlert 阻断告警（推送 归属校管理员+安全团队，响应 ≤5 分钟口径字段）
type CaAlert struct {
	global.GVA_MODEL
	UserName       string `json:"userName" gorm:"type:varchar(100);index;comment:用户"`
	School         string `json:"school" gorm:"type:varchar(100);comment:所属校"`
	Resource       string `json:"resource" gorm:"type:varchar(200);comment:资源"`
	BlockReason    string `json:"blockReason" gorm:"type:varchar(300);comment:阻断原因"`
	PushedTo       string `json:"pushedTo" gorm:"type:varchar(300);comment:推送对象(归属校管理员+安全团队)"`
	ResponseMinutes int   `json:"responseMinutes" gorm:"default:5;comment:响应时限(分钟,≤5 口径)"`
	Handled        bool   `json:"handled" gorm:"index;comment:已处置"`
}

func (CaAlert) TableName() string { return "ca_alert" }
