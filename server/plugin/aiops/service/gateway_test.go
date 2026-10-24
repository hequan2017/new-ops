// Package service aiops LLM 网关错误快照单测
package service

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// TestErrSnippet 非 JSON 响应快照：纯空白长体不越界、CJK 不切半字符、正常短体保留
func TestErrSnippet(t *testing.T) {
	// 原缺陷复现输入：≥201 字节纯空白（TrimSpace 后为空，旧实现 [:200] 越界 panic）
	if got := errSnippet([]byte(strings.Repeat(" \n\t", 100))); got != "" {
		t.Fatalf("纯空白长体应截为空串: %q", got)
	}
	// CJK 边界：200 字节处若为多字节字符中段应回退到字符起点，不产出非法 UTF-8
	cjk := []byte(strings.Repeat("障", 100)) // 每字 3 字节，300 字节
	got := errSnippet(cjk)
	if !utf8.ValidString(got) {
		t.Fatalf("CJK 截断产出非法 UTF-8: %q", got)
	}
	if want := strings.Repeat("障", 66); got != want { // 198 字节=66 字
		t.Fatalf("CJK 截断长度不符: %d 字", len(got)/3)
	}
	// 短体：TrimSpace 生效、内容保留
	if got := errSnippet([]byte("  gateway timeout  ")); got != "gateway timeout" {
		t.Fatalf("短体应去空白保留: %q", got)
	}
}

func TestValidateLLMBaseURL(t *testing.T) {
	for _, c := range []struct {
		url string
		ok  bool
	}{
		{"https://api.openai.com/v1", true},
		{"http://10.0.0.5:11434/v1", true}, // 内网自建网关是核心用法，放行
		{"file:///etc/passwd", false},
		{"gopher://x", false},
		{"https://example.com/v1?token=1", false},
		{"https://example.com/v1#frag", false},
		{"https://user:pass@example.com", false},
		{"notaurl", false},
	} {
		err := validateLLMBaseURL(c.url)
		if c.ok && err != nil {
			t.Fatalf("%q 应合法: %v", c.url, err)
		}
		if !c.ok && err == nil {
			t.Fatalf("%q 应被拒绝", c.url)
		}
	}
}
