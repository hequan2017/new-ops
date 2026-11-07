// Package service llmops 技能鉴权限流单测（令牌哈希校验/令牌桶限流时钟注入）
package service

import (
	"testing"
	"time"
)

func TestGenerateTokenPlain(t *testing.T) {
	tok, err := GenerateTokenPlain()
	if err != nil || len(tok) != 32 {
		t.Fatalf("令牌长度异常: %q %v", tok, err)
	}
	for _, c := range tok {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("应为 hex: %c", c)
		}
	}
	// 两次生成不同
	tok2, _ := GenerateTokenPlain()
	if tok == tok2 {
		t.Fatal("令牌应随机")
	}
}

func TestVerifyTokenHash(t *testing.T) {
	plain, _ := GenerateTokenPlain()
	h := HashToken(plain)
	if !VerifyTokenHash(plain, h) {
		t.Fatal("正确令牌应通过")
	}
	if VerifyTokenHash(plain+"x", h) {
		t.Fatal("篡改令牌应拒绝")
	}
	if !strings_equal(HashToken(plain), h) {
		t.Fatal("哈希应确定")
	}
}

func strings_equal(a, b string) bool { return a == b }

// TestRateLimiter 令牌桶：突发耗尽 → 时钟推进恢复（时钟注入）
func TestRateLimiter(t *testing.T) {
	l := &rateLimiter{buckets: map[string]*tokenBucket{}}
	now := time.Now()
	l.now = func() time.Time { return now }

	// rate=2：首两次放行，第三次拒绝
	if !l.Allow("k", 2) || !l.Allow("k", 2) {
		t.Fatal("初始突发应放行两次")
	}
	if l.Allow("k", 2) {
		t.Fatal("第三次应被限流")
	}
	// 推进 0.5s → 补充 1 个 → 放行
	now = now.Add(500 * time.Millisecond)
	if !l.Allow("k", 2) {
		t.Fatal("时钟推进后应放行")
	}
	if l.Allow("k", 2) {
		t.Fatal("补充不足应再次限流")
	}
	// 不同 key 独立
	if !l.Allow("other", 2) {
		t.Fatal("不同 key 应独立计数")
	}
}
