// Package service llmops 摘要域单测（输入截断/prompt 结构）
package service

import (
	"strings"
	"testing"
)

func TestTruncateForSummary(t *testing.T) {
	// 短文原样
	if got := truncateForSummary("  短文本  ", 100); got != "短文本" {
		t.Fatalf("短文应 trim 原样: %q", got)
	}
	// 长文截断带标注
	long := strings.Repeat("字", 100)
	got := truncateForSummary(long, 50)
	if !strings.HasPrefix(got, strings.Repeat("字", 50)) {
		t.Fatalf("应保留前 50 rune")
	}
	if !strings.HasSuffix(got, "已截取前段生成摘要）") {
		t.Fatalf("截断应带口径标注: %q", got)
	}
	if got := truncateForSummary("", 100); got != "" {
		t.Fatalf("空文应得空串: %q", got)
	}
}

// TestBuildSummaryPrompt 摘要 prompt 契约：system 指令 + user 内容两段式
func TestBuildSummaryPrompt(t *testing.T) {
	msgs := buildSummaryPrompt("文档正文")
	if len(msgs) != 2 || msgs[0].Role != "system" || msgs[1].Role != "user" {
		t.Fatalf("两段式结构异常: %+v", msgs)
	}
	for _, kw := range []string{"摘要", "3-5 句"} {
		if !strings.Contains(msgs[0].Content, kw) {
			t.Fatalf("system 指令缺 %q: %s", kw, msgs[0].Content)
		}
	}
	if msgs[1].Content != "文档正文" {
		t.Fatalf("user 内容透传异常: %s", msgs[1].Content)
	}
}
