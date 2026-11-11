// Package model eduadmin 教育管理一模通管：跨校身份认证接口（需求 3.6-34）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduIdpAdapter 跨校身份源适配器（SAML2.0/OAuth2.0 元数据配置+互认范围）
type EduIdpAdapter struct {
	global.GVA_MODEL
	School           string `json:"school" gorm:"type:varchar(100);uniqueIndex;comment:共建高校"`
	Protocol         string `json:"protocol" gorm:"type:varchar(32);comment:SAML2.0/OAuth2.0"`
	MetadataEndpoint string `json:"metadataEndpoint" gorm:"type:varchar(500);comment:元数据端点"`
	ClientId         string `json:"clientId" gorm:"type:varchar(200);comment:OAuth2.0 ClientId(非机密)"`
	MutualScope      string `json:"mutualScope" gorm:"type:varchar(500);comment:互认范围(教学资源/图书馆/算力等,逗号分隔)"`
	Status           string `json:"status" gorm:"type:varchar(16);comment:启用/停用"`
}

func (EduIdpAdapter) TableName() string { return "edu_idp_adapter" }

// EduIdpUserSync 用户同步任务记录（同步频率 ≤5 分钟口径字段）
type EduIdpUserSync struct {
	global.GVA_MODEL
	School          string `json:"school" gorm:"type:varchar(100);index;comment:高校"`
	SyncFreqMinutes int    `json:"syncFreqMinutes" gorm:"default:5;comment:同步频率(分钟,≤5 口径)"`
	Pulled          int    `json:"pulled" gorm:"comment:拉取用户数"`
	OK              bool   `json:"ok" gorm:"comment:是否成功"`
	Message         string `json:"message" gorm:"type:varchar(500);comment:执行说明"`
}

func (EduIdpUserSync) TableName() string { return "edu_idp_user_sync" }
