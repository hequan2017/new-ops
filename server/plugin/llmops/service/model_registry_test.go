// Package service llmops 模型注册单测（能力标签解析/白名单、服务地址校验）
package service

import (
	"strings"
	"testing"
)

func TestValidCapability(t *testing.T) {
	for _, ok := range []string{"chat", "embed", "rerank"} {
		if !ValidCapability(ok) {
			t.Fatalf("合法能力 %q 被拒绝", ok)
		}
	}
	for _, bad := range []string{"", "Chat", "asr", "vision"} {
		if ValidCapability(bad) {
			t.Fatalf("非法能力 %q 被放行", bad)
		}
	}
}

func TestParseCapabilities(t *testing.T) {
	caps, err := ParseCapabilities(" chat, embed ,CHAT, rerank")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(caps) != 3 || caps[0] != "chat" || caps[1] != "embed" || caps[2] != "rerank" {
		t.Fatalf("去重/小写归一异常: %v", caps)
	}
	if _, err := ParseCapabilities("chat,vision"); err == nil || !strings.Contains(err.Error(), "2228") {
		t.Fatalf("非法标签应报 2228: %v", err)
	}
	if _, err := ParseCapabilities(" , "); err == nil || !strings.Contains(err.Error(), "至少") {
		t.Fatalf("空标签应报错: %v", err)
	}
}

func TestValidateModelBaseURL(t *testing.T) {
	for _, ok := range []string{"http://127.0.0.1:8890/v1", "https://api.deepseek.com/v1"} {
		if err := ValidateModelBaseURL(ok); err != nil {
			t.Fatalf("%q 应放行: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "notaurl", "ftp://x/v1", "file:///tmp"} {
		if err := ValidateModelBaseURL(bad); err == nil {
			t.Fatalf("%q 应拒绝", bad)
		}
	}
}
