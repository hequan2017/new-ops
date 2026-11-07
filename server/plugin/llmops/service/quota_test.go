// Package service llmops 配额单测（跨日重置/余量校验/白名单）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

func TestShouldReset(t *testing.T) {
	if !shouldReset("2026-11-07", "2026-11-08") {
		t.Fatal("跨日应重置")
	}
	if !shouldReset("", "2026-11-08") {
		t.Fatal("空归属日应重置")
	}
	if shouldReset("2026-11-08", "2026-11-08") {
		t.Fatal("同日不应重置")
	}
}

func TestCheckQuota(t *testing.T) {
	// 未超限
	if err := checkQuota(&model.LlmopsQuota{DailyTokenLimit: 100, UsedTokens: 50, DailyCallLimit: 10, UsedCalls: 5}); err != nil {
		t.Fatalf("未超限应放行: %v", err)
	}
	// Token 超限
	err := checkQuota(&model.LlmopsQuota{DailyTokenLimit: 100, UsedTokens: 100, DailyCallLimit: 10, UsedCalls: 0})
	if err == nil || !strings.Contains(err.Error(), "Token") {
		t.Fatalf("Token 超限应拒: %v", err)
	}
	// 调用超限
	err = checkQuota(&model.LlmopsQuota{DailyTokenLimit: -1, UsedTokens: 0, DailyCallLimit: 10, UsedCalls: 10})
	if err == nil || !strings.Contains(err.Error(), "调用") {
		t.Fatalf("调用超限应拒: %v", err)
	}
	// -1 不限
	if err := checkQuota(&model.LlmopsQuota{DailyTokenLimit: -1, UsedTokens: 999999, DailyCallLimit: -1, UsedCalls: 999999}); err != nil {
		t.Fatalf("不限应放行: %v", err)
	}
	// 恰好到达上限（used == limit → 已用尽，下一次拒绝）
	if err := checkQuota(&model.LlmopsQuota{DailyTokenLimit: 10, UsedTokens: 10, DailyCallLimit: -1, UsedCalls: 0}); err == nil {
		t.Fatal("used==limit 应视为用尽")
	}
}

func TestValidQuotaScope(t *testing.T) {
	if !ValidQuotaScope("user") || !ValidQuotaScope("agent") {
		t.Fatal("合法维度应放行")
	}
	if ValidQuotaScope("") || ValidQuotaScope("team") {
		t.Fatal("非法维度应拒绝")
	}
}
