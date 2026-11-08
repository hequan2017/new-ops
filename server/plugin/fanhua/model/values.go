// Package model fanhua 繁花-文科大模型：价值观知识库（需求 3.4-16）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// FanhuaSensitiveWord 敏感词（block 拦截 / replace 替换，对齐 llmops L03 口径）
type FanhuaSensitiveWord struct {
	global.GVA_MODEL
	Word        string `json:"word" gorm:"type:varchar(100);uniqueIndex;comment:敏感词"`
	Action      string `json:"action" gorm:"type:varchar(8);default:block;comment:block/replace"`
	Replacement string `json:"replacement" gorm:"type:varchar(100);comment:替换文案(replace 动作)"`
	Weight      int    `json:"weight" gorm:"default:10;comment:健康度扣减权重"`
	Enabled     bool   `json:"enabled" gorm:"default:true;index;comment:启用"`
}

func (FanhuaSensitiveWord) TableName() string { return "fanhua_sensitive_word" }

// FanhuaComplianceTemplate 合规表达模板
type FanhuaComplianceTemplate struct {
	global.GVA_MODEL
	Name     string `json:"name" gorm:"type:varchar(200);comment:模板名"`
	Scene    string `json:"scene" gorm:"type:varchar(100);comment:适用场景"`
	Template string `json:"template" gorm:"type:text;comment:模板内容"`
}

func (FanhuaComplianceTemplate) TableName() string { return "fanhua_compliance_template" }

// FanhuaConflictCase 多语言价值观冲突案例
type FanhuaConflictCase struct {
	global.GVA_MODEL
	Lang     string `json:"lang" gorm:"type:varchar(16);index;comment:中/英/多语"`
	Content  string `json:"content" gorm:"type:text;comment:案例内容"`
	RiskNote string `json:"riskNote" gorm:"type:varchar(500);comment:风险说明"`
}

func (FanhuaConflictCase) TableName() string { return "fanhua_conflict_case" }
