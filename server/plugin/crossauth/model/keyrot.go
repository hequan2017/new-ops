// Package model crossauth 跨校访问认证与安全：密钥轮换（需求 3.7-10）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CaKeyRotation 密钥轮换表（90 天周期口径，密钥密文落库）
type CaKeyRotation struct {
	global.GVA_MODEL
	KeyName        string `json:"keyName" gorm:"type:varchar(100);uniqueIndex;comment:密钥名"`
	Algorithm      string `json:"algorithm" gorm:"type:varchar(32);comment:SM4-ECB/AES-GCM"`
	SecretEnc      string `json:"-" gorm:"type:varchar(1000);comment:数据密钥密文(AES-GCM 主密钥保护)"`
	CycleDays      int    `json:"cycleDays" gorm:"default:90;comment:轮换周期(天,默认 90)"`
	RotatedAt      string `json:"rotatedAt" gorm:"type:varchar(32);comment:最近轮换时间"`
	NextRotationAt string `json:"nextRotationAt" gorm:"type:varchar(32);comment:下次轮换时间"`
	Status         string `json:"status" gorm:"type:varchar(16);comment:启用/停用"`
}

func (CaKeyRotation) TableName() string { return "ca_key_rotation" }
