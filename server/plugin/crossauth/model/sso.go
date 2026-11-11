// Package model crossauth 跨校访问认证与安全：SSO 单点登录（需求 3.7-4）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CaSsoApp 跨校应用注册
type CaSsoApp struct {
	global.GVA_MODEL
	AppName     string `json:"appName" gorm:"type:varchar(200);uniqueIndex;comment:应用名"`
	AppKey      string `json:"appKey" gorm:"type:varchar(100);uniqueIndex;comment:应用标识"`
	CallbackUrl string `json:"callbackUrl" gorm:"type:varchar(500);comment:回调地址"`
	Status      string `json:"status" gorm:"type:varchar(16);comment:启用/停用"`
}

func (CaSsoApp) TableName() string { return "ca_sso_app" }

// CaSsoTicket 一次性票据（60s 有效、防重放）
type CaSsoTicket struct {
	global.GVA_MODEL
	Ticket    string `json:"ticket" gorm:"type:char(64);uniqueIndex;comment:票据(SHA256 hex)"`
	AppKey    string `json:"appKey" gorm:"type:varchar(100);index;comment:应用标识"`
	UserName  string `json:"userName" gorm:"type:varchar(100);index;comment:用户"`
	Used      bool   `json:"used" gorm:"index;comment:已换取"`
	UsedAt    string `json:"usedAt" gorm:"type:varchar(32);comment:换取时间"`
	ExpiredAt int64  `json:"expiredAt" gorm:"comment:过期时间戳(秒)"`
}

func (CaSsoTicket) TableName() string { return "ca_sso_ticket" }
