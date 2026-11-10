// Package service eduadmin：资产故障预测单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestFaultLevel 阈值+斜率双因子判级
func TestFaultLevel(t *testing.T) {
	cases := []struct {
		threshold, slope, latest float64
		want                     string
	}{
		{100, 0.5, 120, "严重"},  // 达阈值
		{100, 0.5, 90, "轻微"},   // ≥0.8 倍阈值
		{100, 3.0, 50, "轻微"},   // 斜率超容差 2.0
		{100, 1.0, 50, "无故障"}, // 平稳且远离阈值
		{0, 10, 999, "无故障"},   // 阈值未配置
	}
	for _, c := range cases {
		if got := faultLevel(c.threshold, c.slope, c.latest); got != c.want {
			t.Fatalf("阈值 %g/斜率 %g/最新 %g 应 %s，得 %s", c.threshold, c.slope, c.latest, c.want, got)
		}
	}
}

// TestMaintenancePlan 维护计划模板
func TestMaintenancePlan(t *testing.T) {
	if p := maintenancePlan("严重", "中央空调机组"); !strings.Contains(p, "立即停机检修") {
		t.Fatalf("严重计划应含停机: %s", p)
	}
	if p := maintenancePlan("轻微", "电梯主机"); !strings.Contains(p, "7 日内") {
		t.Fatalf("轻微计划应含 7 日巡检: %s", p)
	}
	if p := maintenancePlan("无故障", "实验台"); !strings.Contains(p, "月度保养") {
		t.Fatalf("无故障计划应含月检: %s", p)
	}
}

// TestTrendSlopeViaRegression 时序斜率（复用 linearRegression）
func TestTrendSlopeViaRegression(t *testing.T) {
	x := []float64{0, 1, 2, 3}
	y := []float64{50, 55, 60, 65} // 斜率 5
	slope, _, err := linearRegression(x, y)
	if err != nil || slope != 5 {
		t.Fatalf("斜率应为 5: %v %v", slope, err)
	}
}
