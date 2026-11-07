// Package service exam：实操助手
package service

import (
	"context"
	"strconv"
	"strings"
	"time"
)

// parseUintTarget 目标 ID 解析（operator 通道 target 为算子 ID）
func parseUintTarget(s string) uint {
	v, _ := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
	return uint(v)
}

// contextWithTimeout 实操执行超时（与 llmops 网关口径一致）
func contextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 120*time.Second)
}
