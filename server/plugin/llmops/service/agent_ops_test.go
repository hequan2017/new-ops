// Package service llmops 发布运营单测（快照回写映射/P95/统计聚合）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

func TestSnapshotToUpdates(t *testing.T) {
	snap := `{"name":"机器人","mode":"qa","sysPrompt":"sys","modelId":3,"knowledgeIds":"1,2","maxTurns":8,"configJSON":"{}","ignored":"x"}`
	updates, err := snapshotToUpdates(snap)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	// 列名映射 + 字段白名单
	if updates["name"] != "机器人" || updates["sys_prompt"] != "sys" || updates["model_id"] != float64(3) {
		t.Fatalf("映射异常: %+v", updates)
	}
	if updates["max_turns"] != float64(8) || updates["config_json"] != "{}" {
		t.Fatalf("映射异常2: %+v", updates)
	}
	if _, ok := updates["ignored"]; ok {
		t.Fatal("白名单外字段不应回写")
	}
	// 坏快照
	if _, err := snapshotToUpdates("bad"); err == nil {
		t.Fatal("坏快照应拒绝")
	}
}

func TestPercentile95(t *testing.T) {
	if got := percentile95(nil); got != 0 {
		t.Fatalf("空应 0: %v", got)
	}
	if got := percentile95([]int64{100}); got != 100 {
		t.Fatalf("单值: %v", got)
	}
	// 0..99 共 100 个 → P95 = 第 94 位（升序）= 94
	var lat []int64
	for i := 0; i < 100; i++ {
		lat = append(lat, int64(i))
	}
	if got := percentile95(lat); got != 94 {
		t.Fatalf("P95 异常: %v", got)
	}
}

func TestSummarizeCalls(t *testing.T) {
	mk := func(ok bool, ms int64) model.LlmopsAgentCallLog {
		r := model.LlmopsAgentCallLog{LatencyMs: ms, Ok: ok}
		return r
	}
	rows := []model.LlmopsAgentCallLog{mk(true, 100), mk(true, 200), mk(false, 1000)}
	stat := summarizeCalls(rows)
	if stat.Total != 3 || stat.OkCount != 2 || stat.FailCount != 1 || stat.AvgLatency != 433 {
		t.Fatalf("统计异常: %+v", stat)
	}
	if stat.P95Latency != 1000 {
		t.Fatalf("P95 应取最大附近: %v", stat.P95Latency)
	}
	// 空集
	empty := summarizeCalls(nil)
	if empty.Total != 0 || empty.AvgLatency != 0 || empty.P95Latency != 0 {
		t.Fatalf("空集统计异常: %+v", empty)
	}
}
