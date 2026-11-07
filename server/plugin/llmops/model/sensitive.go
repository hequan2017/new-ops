// Package model llmops：敏感词（要求(8) 干预机制词库）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// LlmopsSensitiveWord 敏感词（block 拦截 / replace 替换放行）
type LlmopsSensitiveWord struct {
	global.GVA_MODEL
	Word        string `json:"word" gorm:"type:varchar(64);uniqueIndex;comment:敏感词"`
	Action      string `json:"action" gorm:"type:varchar(16);default:block;comment:block/replace"`
	Replacement string `json:"replacement" gorm:"type:varchar(64);comment:替换词(action=replace)"`
	Enabled     bool   `json:"enabled" gorm:"default:true;comment:启用"`
}

func (LlmopsSensitiveWord) TableName() string { return "llmops_sensitive_word" }
