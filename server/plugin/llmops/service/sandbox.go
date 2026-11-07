// Package service llmops：技能执行沙箱（要求(6) 沙箱隔离）
// 口径：四道护栏——超时（goroutine+select 强制，不依赖工具自觉使用 ctx）、
// 结果截断、URL 白名单（实现 endpointURLer 接口的工具受检）、敏感参数脱敏（trace 展示用）。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// 错误码（llmops 段续用）
const ErrCodeSandbox = 2291

// SandboxConfig 沙箱配置（零值取默认）
type SandboxConfig struct {
	Timeout         time.Duration
	MaxResultRunes  int
	AllowedURLHosts []string // 空=不限（http/https）
	SensitiveKeys   []string // 空=默认敏感键
}

// 默认敏感键（大小写不敏感）
var defaultSensitiveKeys = []string{"password", "passwd", "secret", "token", "apikey", "api_key", "authorization", "authvalue"}

// endpointURLer 带 URL 的工具接口（MCP 工具实现，受 URL 白名单约束）
type endpointURLer interface {
	EndpointURL() string
}

// MaskSensitiveValue 敏感值掩码（保留首尾各 2 字符，纯函数）
func MaskSensitiveValue(v string) string {
	r := []rune(v)
	if len(r) <= 4 {
		return strings.Repeat("*", len(r))
	}
	return string(r[:2]) + strings.Repeat("*", len(r)-4) + string(r[len(r)-2:])
}

// SanitizeArgs 参数脱敏（纯函数，测试锚点）：敏感键值掩码（含一层嵌套对象）
func SanitizeArgs(argsJSON string, sensitiveKeys []string) string {
	if strings.TrimSpace(argsJSON) == "" {
		return argsJSON
	}
	keys := sensitiveKeys
	if len(keys) == 0 {
		keys = defaultSensitiveKeys
	}
	sensitive := map[string]bool{}
	for _, k := range keys {
		sensitive[strings.ToLower(k)] = true
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &obj); err != nil {
		return argsJSON // 非 JSON 原样（展示层兜底）
	}
	var mask func(m map[string]any)
	mask = func(m map[string]any) {
		for k, v := range m {
			if sensitive[strings.ToLower(k)] {
				if s, ok := v.(string); ok {
					m[k] = MaskSensitiveValue(s)
				}
				continue
			}
			if nested, ok := v.(map[string]any); ok {
				mask(nested)
			}
		}
	}
	mask(obj)
	b, err := json.Marshal(obj)
	if err != nil {
		return argsJSON
	}
	return string(b)
}

// ValidateURLWhitelist URL 白名单校验（纯函数：allowed 空=不限 http/https）
func ValidateURLWhitelist(rawURL string, allowed []string) error {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Host == "" {
		return fmt.Errorf("[%d]URL 无法解析: %s", ErrCodeSandbox, rawURL)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("[%d]仅允许 http/https", ErrCodeSandbox)
	}
	if len(allowed) == 0 {
		return nil
	}
	host := strings.ToLower(u.Hostname())
	for _, a := range allowed {
		a = strings.ToLower(strings.TrimSpace(a))
		if a == "" {
			continue
		}
		if host == a || strings.HasSuffix(host, "."+a) {
			return nil
		}
	}
	return fmt.Errorf("[%d]目标主机不在白名单: %s", ErrCodeSandbox, host)
}

// RunToolSandboxed 沙箱执行：脱敏参数回传 + 超时/截断/白名单护栏
func RunToolSandboxed(tool AgentTool, argsJSON string, cfg SandboxConfig) (maskedArgs, result string, err error) {
	maskedArgs = SanitizeArgs(argsJSON, cfg.SensitiveKeys)
	// URL 白名单（实现 endpointURLer 的工具）
	if ep, ok := tool.(endpointURLer); ok {
		if err := ValidateURLWhitelist(ep.EndpointURL(), cfg.AllowedURLHosts); err != nil {
			return maskedArgs, "", err
		}
	}
	// 超时护栏（goroutine+select，不依赖工具使用 ctx）
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	type outcome struct {
		res string
		err error
	}
	done := make(chan outcome, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- outcome{err: fmt.Errorf("[%d]工具执行 panic: %v", ErrCodeSandbox, r)}
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		res, err := tool.Execute(ctx, argsJSON)
		done <- outcome{res: res, err: err}
	}()
	select {
	case o := <-done:
		if o.err != nil {
			return maskedArgs, "", o.err
		}
		result = o.res
	case <-time.After(timeout): // 强制超时（工具不响应 ctx 也生效；done 带缓冲不阻塞泄漏协程）
		return maskedArgs, "", fmt.Errorf("[%d]工具执行超时(>%s)", ErrCodeSandbox, timeout)
	}
	maxRunes := cfg.MaxResultRunes
	if maxRunes <= 0 {
		maxRunes = 4000
	}
	if r := []rune(result); len(r) > maxRunes {
		result = string(r[:maxRunes]) + "…(沙箱截断)"
	}
	return maskedArgs, result, nil
}
