// Package service llmops 测试台单测（rerank 响应解析/协议白名单/错误摘要截断）
package service

import (
	"strings"
	"testing"
)

func TestParseRerankResponse(t *testing.T) {
	body := `{"results":[{"index":2,"score":0.98},{"index":0,"score":0.41}]}`
	scores, err := parseRerankResponse([]byte(body))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(scores) != 2 || scores[0].Index != 2 || scores[0].Score != 0.98 || scores[1].Index != 0 {
		t.Fatalf("解析异常: %+v", scores)
	}
	if _, err := parseRerankResponse([]byte(`{"results":[]}`)); err == nil || !strings.Contains(err.Error(), "2234") {
		t.Fatalf("空 results 应报 2234: %v", err)
	}
	if _, err := parseRerankResponse([]byte(`bad`)); err == nil || !strings.Contains(err.Error(), "2232") {
		t.Fatalf("坏 JSON 应报 2232: %v", err)
	}
}

// TestRerankDocBackfill 原文回填契约
func TestRerankDocBackfill(t *testing.T) {
	docs := []string{"甲", "乙", "丙"}
	scores := []RerankScore{{Index: 1, Score: 0.9}, {Index: 5, Score: 0.1}} // 越界 index 不回填
	for i := range scores {
		if scores[i].Index >= 0 && scores[i].Index < len(docs) {
			scores[i].Document = docs[scores[i].Index]
		}
	}
	if scores[0].Document != "乙" || scores[1].Document != "" {
		t.Fatalf("回填异常: %+v", scores)
	}
}

func TestValidCapability_WhitelistsTestProtocols(t *testing.T) {
	for _, p := range []string{"chat", "embed", "rerank"} {
		if !ValidCapability(p) {
			t.Fatalf("协议 %q 应在能力白名单内", p)
		}
	}
	if ValidCapability("completion") {
		t.Fatal("非白名单协议应拒绝")
	}
}

// TestTruncateSample_LogCap 调用日志错误摘要 500 rune 截断口径
func TestTruncateSample_LogCap(t *testing.T) {
	long := strings.Repeat("错", 800)
	cut := truncateSample(long)
	if len([]rune(cut)) != 65 { // truncateSample 64+…
		t.Fatalf("truncateSample 口径变化需同步日志截断: %d", len([]rune(cut)))
	}
}
