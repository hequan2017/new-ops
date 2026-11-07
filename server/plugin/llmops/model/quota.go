// Package model llmops：资源配额（要求(8) 资源管理）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// LlmopsQuota 配额（scope=user/agent；计数跨日自动重置）
type LlmopsQuota struct {
	global.GVA_MODEL
	Scope           string `json:"scope" gorm:"type:varchar(8);index;comment:user/agent"`
	ScopeKey        string `json:"scopeKey" gorm:"type:varchar(64);uniqueIndex:uk_scope;comment:用户ID或智能体ID"`
	DailyTokenLimit int64  `json:"dailyTokenLimit" gorm:"default:1000000;comment:日 Token 上限"`
	DailyCallLimit  int64  `json:"dailyCallLimit" gorm:"default:1000;comment:日调用上限"`
	UsedTokens      int64  `json:"usedTokens" gorm:"comment:今日已用 Token"`
	UsedCalls       int64  `json:"usedCalls" gorm:"comment:今日已用调用"`
	ResetDate       string `json:"resetDate" gorm:"type:varchar(10);comment:计数归属日(YYYY-MM-DD)"`
}

func (LlmopsQuota) TableName() string { return "llmops_quota" }
