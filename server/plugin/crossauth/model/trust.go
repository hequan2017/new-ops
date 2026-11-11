// Package model crossauth 跨校访问认证与安全：身份互认规则库（需求 3.7-2）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CaTrustRule 互认规则（本校身份 × 目标资源 × 互认范围）
type CaTrustRule struct {
	global.GVA_MODEL
	HomeIdentity  string `json:"homeIdentity" gorm:"type:varchar(100);index;comment:本校身份(如 上海大学教师)"`
	TargetResource string `json:"targetResource" gorm:"type:varchar(200);index;comment:目标资源(如 合作高校教学资源库)"`
	Scope         string `json:"scope" gorm:"type:varchar(300);comment:互认范围(逗号分隔)"`
	Allowed       bool   `json:"allowed" gorm:"comment:是否允许"`
	Remark        string `json:"remark" gorm:"type:varchar(300);comment:备注"`
}

func (CaTrustRule) TableName() string { return "ca_trust_rule" }
