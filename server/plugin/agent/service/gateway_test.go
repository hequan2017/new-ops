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

// TestRegistryDisplaceAndUnregister 同资产重连顶替与退出清理：
// 旧连接退出不得误删顶替它的新连接注册项（网络闪断重连场景）
func TestRegistryDisplaceAndUnregister(t *testing.T) {
	const asset = uint(42)
	old, fresh := &agentConn{AssetID: asset}, &agentConn{AssetID: asset}
	defer unregisterConn(asset, nil) // 兜底清场（nil 不等于任何登记项，无副作用）

	if displaced := registerConn(asset, old); displaced != nil {
		t.Fatalf("首次登记不应有顶替对象")
	}
	// 重连：新连接顶替旧连接
	if displaced := registerConn(asset, fresh); displaced != old {
		t.Fatalf("重连应返回被顶替的旧连接")
	}
	// 旧连接读循环退出、执行清理——新连接的注册项必须仍在
	unregisterConn(asset, old)
	registry.Lock()
	_, ok := registry.conns[asset]
	registry.Unlock()
	if !ok {
		t.Fatalf("旧连接退出误删了新连接的注册项")
	}
	// 新连接退出才真正摘除
	unregisterConn(asset, fresh)
	registry.Lock()
	_, ok = registry.conns[asset]
	registry.Unlock()
	if ok {
		t.Fatalf("新连接退出后注册表应清空")
	}
}
