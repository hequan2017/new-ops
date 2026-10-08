// Package service Agent 网关纯逻辑单测
package service

import (
	"testing"
	"time"
)

func TestAgentStatusByHeartbeat(t *testing.T) {
	now := time.Now()
	fresh := now.Add(-AgentHeartbeatInterval)
	stale := now.Add(-AgentOnlineTTL - time.Second)
	if s := AgentStatusByHeartbeat(&fresh, now); s != "在线" {
		t.Fatalf("新鲜心跳应为在线: %s", s)
	}
	if s := AgentStatusByHeartbeat(&stale, now); s != "离线" {
		t.Fatalf("过期心跳应为离线: %s", s)
	}
	if s := AgentStatusByHeartbeat(nil, now); s != "离线" {
		t.Fatalf("无心跳应为离线: %s", s)
	}
}

func TestSanitizeMetricPoints(t *testing.T) {
	out := sanitizeMetricPoints([]agentPoint{
		{Name: "cpu_percent", Value: 12.34},
		{Name: " MEM_PERCENT ", Value: 55},
		{Name: "hack_metric", Value: 1}, // 白名单外
		{Name: "load1", Value: 1e15},    // 越界
	})
	if len(out) != 2 {
		t.Fatalf("应保留 2 项: %+v", out)
	}
	if out[0].Name != "cpu_percent" || out[1].Name != "mem_percent" {
		t.Fatalf("名称归一错误: %+v", out)
	}
	if len(sanitizeMetricPoints(nil)) != 0 {
		t.Fatal("空输入应返回空")
	}
}

func TestHashAgentToken(t *testing.T) {
	a := HashAgentToken("token-a")
	b := HashAgentToken("token-b")
	if a == b || len(a) != 64 {
		t.Fatalf("哈希异常: %s vs %s", a, b)
	}
	if HashAgentToken("token-a") != a {
		t.Fatal("同输入应同哈希")
	}
}
