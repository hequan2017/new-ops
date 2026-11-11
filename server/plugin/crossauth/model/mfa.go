// Package model crossauth 跨校访问认证与安全：MFA 管理（需求 3.7-3）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CaMfaConfig 用户 MFA 配置（方式白名单；APP 动态码持有 TOTP 密钥密文）
type CaMfaConfig struct {
	global.GVA_MODEL
	UserName   string `json:"userName" gorm:"type:varchar(100);uniqueIndex;comment:用户"`
	Method     string `json:"method" gorm:"type:varchar(32);comment:短信验证码/APP动态码/指纹/人脸"`
	TotpSecretEnc string `json:"-" gorm:"type:varchar(1000);comment:TOTP 密钥密文(AES-GCM)"`
	Status     string `json:"status" gorm:"type:varchar(16);index;comment:启用/停用"`
}

func (CaMfaConfig) TableName() string { return "ca_mfa_config" }

// CaMfaPolicy 高权限角色强制 MFA 策略
type CaMfaPolicy struct {
	global.GVA_MODEL
	RoleName string `json:"roleName" gorm:"type:varchar(100);uniqueIndex;comment:高权限角色"`
	ForceMfa bool   `json:"forceMfa" gorm:"comment:是否强制 MFA"`
	Remark   string `json:"remark" gorm:"type:varchar(300);comment:备注"`
}

func (CaMfaPolicy) TableName() string { return "ca_mfa_policy" }
