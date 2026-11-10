// Package service eduadmin：能耗优化单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// TestAggregateEnergy 楼栋聚合+夜间占比+同比异常
func TestAggregateEnergy(t *testing.T) {
	current := []model.EduEnergyRecord{
		{Building: "A栋", Year: 2026, Month: 6, Period: "白天", Kwh: 800},
		{Building: "A栋", Year: 2026, Month: 6, Period: "夜间", Kwh: 200},
		{Building: "B栋", Year: 2026, Month: 6, Period: "白天", Kwh: 500},
	}
	last := []model.EduEnergyRecord{
		{Building: "A栋", Year: 2025, Month: 6, Period: "白天", Kwh: 1000},
		{Building: "B栋", Year: 2025, Month: 6, Period: "白天", Kwh: 400},
	}
	out := aggregateEnergy(current, last)
	if len(out) != 2 {
		t.Fatalf("应 2 楼栋: %+v", out)
	}
	// A 栋总 1000 vs 1000 → 0 涨幅；B 栋 500 vs 400 → +25% 异常
	if out[0].Building != "A栋" {
		t.Fatalf("总耗能降序 A 栋应居首: %+v", out)
	}
	b := out[1]
	if b.Building != "B栋" || !b.Abnormal || b.YoYGrowth != 0.25 {
		t.Fatalf("B 栋应 +25%% 异常: %+v", b)
	}
	a := out[0]
	if a.Abnormal || a.NightRate != 0.2 {
		t.Fatalf("A 栋应平稳且夜间占比 0.2: %+v", a)
	}
}

// TestEnergySuggestions 建议模板
func TestEnergySuggestions(t *testing.T) {
	usages := []buildingUsage{
		{Building: "A栋", TotalKwh: 1000, NightRate: 0.2},
		{Building: "B栋", TotalKwh: 900, NightRate: 0.35, YoYGrowth: 0.3, Abnormal: true},
	}
	out := energySuggestions(usages)
	joined := strings.Join(out, "|")
	if !strings.Contains(joined, "B栋】同比上涨 30%") || !strings.Contains(joined, "夜间用电占比 35%") {
		t.Fatalf("应含异常与夜间建议: %s", joined)
	}
	if !strings.Contains(joined, "A栋】为当前最大耗能楼栋") {
		t.Fatalf("最大耗能楼栋应为 A 栋: %s", joined)
	}
	if out := energySuggestions(nil); len(out) != 1 {
		t.Fatalf("空输入应兜底: %+v", out)
	}
}
