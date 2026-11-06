// Package service llmops 多模型调度单测（时延择优纯函数 + DefaultFor 优先口径）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

func mkCandidate(id uint, defaultFor string) model.LlmopsModel {
	m := model.LlmopsModel{Capability: "chat", DefaultFor: defaultFor, Status: "enabled"}
	m.ID = id
	return m
}

// TestPickByLatency 无记录优先于有记录；同分取小 ID
func TestPickByLatency(t *testing.T) {
	candidates := []model.LlmopsModel{mkCandidate(1, ""), mkCandidate(2, ""), mkCandidate(3, "")}
	latency := map[uint]int64{1: 200, 2: 0, 3: 50}
	// 无记录的 2 视为 0 → 最优
	if got := PickByLatency(candidates, latency); got != 2 {
		t.Fatalf("无记录应优先: got %d", got)
	}
	// 全有记录：最低均值得选
	latency2 := map[uint]int64{1: 300, 2: 150, 3: 150}
	if got := PickByLatency(candidates, latency2); got != 2 {
		t.Fatalf("同分应取小 ID: got %d", got)
	}
	// 空 map：全无记录 → 最小 ID
	if got := PickByLatency(candidates, nil); got != 1 {
		t.Fatalf("全无记录应取最小 ID: got %d", got)
	}
}

// TestPickByLatency_Single 单候选
func TestPickByLatency_Single(t *testing.T) {
	if got := PickByLatency([]model.LlmopsModel{mkCandidate(9, "")}, nil); got != 9 {
		t.Fatalf("单候选应直接命中: %d", got)
	}
}

// TestSelectModel_CapabilityGuard 白名单守卫（不触 DB）
func TestSelectModel_CapabilityGuard(t *testing.T) {
	svc := &LlmopsServiceStruct{}
	if _, err := svc.SelectModel("vision"); err == nil {
		t.Fatal("非白名单能力应拒绝")
	}
}
