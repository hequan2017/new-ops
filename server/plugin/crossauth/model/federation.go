// Package model crossauth 跨校访问认证与安全：身份系统对接（需求 3.7-1）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CaFederation 高校身份源对接（协议白名单，密钥 AES-GCM 密文落库）
type CaFederation struct {
	global.GVA_MODEL
	School    string `json:"school" gorm:"type:varchar(100);uniqueIndex;comment:高校"`
	Protocol  string `json:"protocol" gorm:"type:varchar(32);comment:SAML2.0/OAuth2.0"`
	Endpoint  string `json:"endpoint" gorm:"type:varchar(500);comment:用户拉取端点"`
	SecretEnc string `json:"-" gorm:"type:varchar(1000);comment:接口密钥密文(AES-GCM)"`
	Status    string `json:"status" gorm:"type:varchar(16);comment:启用/停用"`
}

func (CaFederation) TableName() string { return "ca_federation" }

// CaSyncUser 同步用户（身份状态白名单）
type CaSyncUser struct {
	global.GVA_MODEL
	Name           string `json:"name" gorm:"type:varchar(100);comment:姓名"`
	UserNo         string `json:"userNo" gorm:"type:varchar(64);uniqueIndex:idx_ca_school_no;comment:学号/工号"`
	School         string `json:"school" gorm:"type:varchar(100);uniqueIndex:idx_ca_school_no;index;comment:院校"`
	Dept           string `json:"dept" gorm:"type:varchar(100);comment:部门/院系"`
	IdentityStatus string `json:"identityStatus" gorm:"type:varchar(16);index;comment:在读/在职/离校/离职"`
}

func (CaSyncUser) TableName() string { return "ca_sync_user" }

// CaSyncTask 同步任务执行记录（频率 ≤5 分钟口径字段）
type CaSyncTask struct {
	global.GVA_MODEL
	FederationId uint   `json:"federationId" gorm:"index;comment:对接配置ID"`
	FreqMinutes  int    `json:"freqMinutes" gorm:"default:5;comment:同步频率(分钟,≤5 口径)"`
	Pulled       int    `json:"pulled" gorm:"comment:拉取用户数"`
	Inserted     int    `json:"inserted" gorm:"comment:新入库数"`
	Updated      int    `json:"updated" gorm:"comment:状态更新数"`
	OK           bool   `json:"ok" gorm:"comment:是否成功"`
	Message      string `json:"message" gorm:"type:varchar(500);comment:执行说明"`
}

func (CaSyncTask) TableName() string { return "ca_sync_task" }
