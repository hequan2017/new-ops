// Package service llmops：资源配额（要求(8)）——Token/调用双维日配额 + 行锁扣减
// 口径：计数跨日自动重置（ResetDate 非当日先清零再扣）；扣减事务行锁防并发超限；
// 上限 -1 表示不限；行锁口径对齐 I04/I05。
package service

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const ErrCodeQuota = 2299

// ValidQuotaScope 配额维度白名单
func ValidQuotaScope(scope string) bool {
	return scope == "user" || scope == "agent"
}

// shouldReset 跨日重置判定（纯函数，测试锚点）
func shouldReset(resetDate, today string) bool {
	return resetDate != today
}

// checkQuota 余量校验（纯函数：-1 不限）
func checkQuota(q *model.LlmopsQuota) error {
	if q.DailyTokenLimit >= 0 && q.UsedTokens >= q.DailyTokenLimit {
		return fmt.Errorf("[%d]Token 日配额已用尽（%d/%d）", ErrCodeQuota, q.UsedTokens, q.DailyTokenLimit)
	}
	if q.DailyCallLimit >= 0 && q.UsedCalls >= q.DailyCallLimit {
		return fmt.Errorf("[%d]调用日配额已用尽（%d/%d）", ErrCodeQuota, q.UsedCalls, q.DailyCallLimit)
	}
	return nil
}

// today 計數归属日
func today() string { return time.Now().Format("2006-01-02") }

// SetQuota 设置配额（管理面 upsert）
func (s *LlmopsServiceStruct) SetQuota(scope, scopeKey string, tokenLimit, callLimit int64) (*model.LlmopsQuota, error) {
	if !ValidQuotaScope(scope) || strings.TrimSpace(scopeKey) == "" {
		return nil, fmt.Errorf("[%d]维度须为 user/agent 且键必填", ErrCodeQuota)
	}
	if tokenLimit < -1 || callLimit < -1 {
		return nil, fmt.Errorf("[%d]上限非法（-1 表示不限）", ErrCodeQuota)
	}
	var q model.LlmopsQuota
	err := global.GVA_DB.Where("scope = ? AND scope_key = ?", scope, scopeKey).First(&q).Error
	if err == gorm.ErrRecordNotFound {
		q = model.LlmopsQuota{Scope: scope, ScopeKey: scopeKey,
			DailyTokenLimit: tokenLimit, DailyCallLimit: callLimit, ResetDate: today()}
		if err := global.GVA_DB.Create(&q).Error; err != nil {
			return nil, err
		}
		return &q, nil
	}
	if err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Model(&q).Updates(map[string]any{
		"daily_token_limit": tokenLimit, "daily_call_limit": callLimit,
	}).Error; err != nil {
		return nil, err
	}
	q.DailyTokenLimit, q.DailyCallLimit = tokenLimit, callLimit
	return &q, nil
}

// GetQuotas 配额列表
func (s *LlmopsServiceStruct) GetQuotas(scope string) ([]model.LlmopsQuota, error) {
	db := global.GVA_DB.Model(&model.LlmopsQuota{})
	if scope != "" {
		if !ValidQuotaScope(scope) {
			return nil, fmt.Errorf("[%d]不支持的维度: %s", ErrCodeQuota, scope)
		}
		db = db.Where("scope = ?", scope)
	}
	var list []model.LlmopsQuota
	err := db.Order("id").Find(&list).Error
	return list, err
}

// ConsumeQuota 扣减（行锁事务：跨日先重置；超限拒并回滚）
func (s *LlmopsServiceStruct) ConsumeQuota(scope, scopeKey string, tokens, calls int64) (*model.LlmopsQuota, error) {
	if !ValidQuotaScope(scope) {
		return nil, fmt.Errorf("[%d]维度须为 user/agent", ErrCodeQuota)
	}
	var q model.LlmopsQuota
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// 行锁取行（无则建默认档）
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("scope = ? AND scope_key = ?", scope, scopeKey).First(&q).Error
		if err == gorm.ErrRecordNotFound {
			q = model.LlmopsQuota{Scope: scope, ScopeKey: scopeKey,
				DailyTokenLimit: 1000000, DailyCallLimit: 1000, ResetDate: today()}
			if err := tx.Create(&q).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		// 跨日重置
		if shouldReset(q.ResetDate, today()) {
			q.UsedTokens, q.UsedCalls, q.ResetDate = 0, 0, today()
		}
		if err := checkQuota(&q); err != nil {
			return err
		}
		q.UsedTokens += tokens
		q.UsedCalls += calls
		return tx.Model(&q).Updates(map[string]any{
			"used_tokens": q.UsedTokens, "used_calls": q.UsedCalls, "reset_date": q.ResetDate,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// QuotaIn 配额设置请求
type QuotaIn struct {
	Scope           string `json:"scope" binding:"required"`
	ScopeKey        string `json:"scopeKey" binding:"required"`
	DailyTokenLimit int64  `json:"dailyTokenLimit"`
	DailyCallLimit  int64  `json:"dailyCallLimit"`
}

// QuotaConsumeIn 配额扣减请求
type QuotaConsumeIn struct {
	Scope    string `json:"scope" binding:"required"`
	ScopeKey string `json:"scopeKey" binding:"required"`
	Tokens   int64  `json:"tokens"`
	Calls    int64  `json:"calls"`
}
