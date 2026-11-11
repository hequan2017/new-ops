// Package model crossauth 跨校访问认证与安全：权限生命周期（需求 3.7-8）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CaGrantExpiry 授权有效期（到期自动回收）
type CaGrantExpiry struct {
	global.GVA_MODEL
	UserName    string `json:"userName" gorm:"type:varchar(100);index;comment:用户"`
	RoleName    string `json:"roleName" gorm:"type:varchar(100);comment:角色"`
	School      string `json:"school" gorm:"type:varchar(100);comment:院校"`
	ExpiredAt   string `json:"expiredAt" gorm:"type:varchar(16);index;comment:到期日(yyyy-MM-dd)"`
	Revoked     bool   `json:"revoked" gorm:"index;comment:已回收"`
	RevokedAt   string `json:"revokedAt" gorm:"type:varchar(32);comment:回收时间"`
}

func (CaGrantExpiry) TableName() string { return "ca_grant_expiry" }

// CaExpiryReminder 到期提醒（到期前 7 天，按授权幂等）
type CaExpiryReminder struct {
	global.GVA_MODEL
	UserName   string `json:"userName" gorm:"type:varchar(100);uniqueIndex:idx_ca_remind_user_role;comment:用户"`
	RoleName   string `json:"roleName" gorm:"type:varchar(100);uniqueIndex:idx_ca_remind_user_role;comment:角色"`
	DaysLeft   int    `json:"daysLeft" gorm:"comment:剩余天数"`
	RemindedAt string `json:"remindedAt" gorm:"type:varchar(32);comment:提醒时间"`
}

func (CaExpiryReminder) TableName() string { return "ca_expiry_reminder" }
