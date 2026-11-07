// Package service llmops 沙箱单测（敏感值掩码/参数脱敏/URL 白名单/超时与截断护栏）
package service

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestMaskSensitiveValue(t *testing.T) {
	if got := MaskSensitiveValue("abcd1234wxyz"); got != "ab********yz" {
		t.Fatalf("掩码异常: %q", got)
	}
	if got := MaskSensitiveValue("abc"); got != "***" {
		t.Fatalf("短值应全掩: %q", got)
	}
}

func TestSanitizeArgs(t *testing.T) {
	// 默认敏感键（大小写不敏感）
	got := SanitizeArgs(`{"user":"张三","Password":"abcd1234wxyz","Token":"secret-token-xyz"}`, nil)
	if !strings.Contains(got, `"user":"张三"`) {
		t.Fatalf("非敏感字段应原样: %s", got)
	}
	if strings.Contains(got, "abcd1234wxyz") || strings.Contains(got, "secret-token-xyz") {
		t.Fatalf("敏感值应掩码: %s", got)
	}
	// 一层嵌套
	got = SanitizeArgs(`{"db":{"password":"topsecret99"},"k":1}`, nil)
	if strings.Contains(got, "topsecret99") {
		t.Fatalf("嵌套敏感键应掩码: %s", got)
	}
	// 非 JSON 原样
	if got := SanitizeArgs("plain text", nil); got != "plain text" {
		t.Fatalf("非 JSON 应原样: %q", got)
	}
}

func TestValidateURLWhitelist(t *testing.T) {
	// 空=不限
	if err := ValidateURLWhitelist("http://any.host/api", nil); err != nil {
		t.Fatalf("空白名单应不限: %v", err)
	}
	allowed := []string{"erp.corp.com"}
	if err := ValidateURLWhitelist("https://erp.corp.com/api/x", allowed); err != nil {
		t.Fatalf("精确命中应放行: %v", err)
	}
	if err := ValidateURLWhitelist("https://api.erp.corp.com/x", allowed); err != nil {
		t.Fatalf("子域应放行: %v", err)
	}
	// 子串伪装拒绝（evil-erp.corp.com 不含 . 前缀）
	if err := ValidateURLWhitelist("https://evil-erp.corp.com/x", allowed); err == nil {
		t.Fatal("伪装域应拒绝")
	}
	if err := ValidateURLWhitelist("https://other.com/x", allowed); err == nil {
		t.Fatal("白名单外应拒绝")
	}
	if err := ValidateURLWhitelist("file:///etc/passwd", nil); err == nil {
		t.Fatal("非 http 协议应拒绝")
	}
}

// slowTool 慢工具（超时护栏验证）
type slowTool struct{ delay time.Duration }

func (slowTool) Name() string             { return "slow" }
func (slowTool) Description() string      { return "slow" }
func (t slowTool) Execute(context.Context, string) (string, error) {
	time.Sleep(t.delay)
	return "done", nil
}

// maskedTool 实现 endpointURLer
type maskedTool struct{ endpoint string }

func (maskedTool) Name() string        { return "masked" }
func (maskedTool) Description() string { return "m" }
func (m maskedTool) Execute(_ context.Context, _ string) (string, error) {
	return strings.Repeat("字", 100), nil
}
func (m maskedTool) EndpointURL() string { return m.endpoint }

func TestRunToolSandboxed(t *testing.T) {
	// 超时：工具睡 200ms，沙箱 50ms
	_, _, err := RunToolSandboxed(slowTool{delay: 200 * time.Millisecond}, `{}`, SandboxConfig{Timeout: 50 * time.Millisecond})
	if err == nil || !strings.Contains(err.Error(), "2291") {
		t.Fatalf("超时应报 2291: %v", err)
	}
	// URL 白名单拦截
	_, _, err = RunToolSandboxed(maskedTool{endpoint: "http://evil.com/x"}, `{}`, SandboxConfig{AllowedURLHosts: []string{"erp.corp.com"}})
	if err == nil || !strings.Contains(err.Error(), "白名单") {
		t.Fatalf("白名单外应拒: %v", err)
	}
	// 结果截断
	_, result, err := RunToolSandboxed(maskedTool{endpoint: "http://ok.com"}, `{}`, SandboxConfig{MaxResultRunes: 20})
	if err != nil || !strings.HasSuffix(result, "…(沙箱截断)") {
		t.Fatalf("截断异常: %q %v", result, err)
	}
	// 参数脱敏回传
	masked, _, _ := RunToolSandboxed(slowTool{delay: time.Millisecond}, `{"token":"abcd1234wxyz"}`, SandboxConfig{Timeout: time.Second})
	if strings.Contains(masked, "abcd1234wxyz") {
		t.Fatalf("脱敏回传异常: %s", masked)
	}
}
