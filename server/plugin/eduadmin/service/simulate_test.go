// Package service eduadmin：政策模拟单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestSimulateFinance 弹性系数模型
func TestSimulateFinance(t *testing.T) {
	in := SimInput{Name: "方案A", BaseBudget: 1000, BaseRevenue: 500, BudgetChangeRate: -0.08, FeeRateChange: 0.1, Elasticity: 0.6}
	res, err := simulateFinance(in)
	if err != nil {
		t.Fatalf("模拟失败: %v", err)
	}
	if res.DeltaBudget != -80 {
		t.Fatalf("预算影响应 -80: %v", res.DeltaBudget)
	}
	if res.DeltaRevenue != 30 { // 500*0.1*0.6
		t.Fatalf("费率收入应 +30: %v", res.DeltaRevenue)
	}
	if res.Net != -50 || res.NetRate != -0.05 {
		t.Fatalf("净影响应 -50(-5%%): %v %v", res.Net, res.NetRate)
	}
	// -5% 恰好为中（< -5% 才高）
	if res.RiskLevel != "中" {
		t.Fatalf("-5%% 应为中: %s", res.RiskLevel)
	}
	// 缺省弹性 0.6
	in2 := SimInput{BaseBudget: 1000, BaseRevenue: 500, BudgetChangeRate: 0.1, FeeRateChange: 0.2}
	res2, _ := simulateFinance(in2)
	if res2.DeltaRevenue != 60 {
		t.Fatalf("缺省弹性 0.6 应 +60: %v", res2.DeltaRevenue)
	}
	if res2.RiskLevel != "低" {
		t.Fatalf("净增应低: %s", res2.RiskLevel)
	}
	// 高风险
	in3 := SimInput{BaseBudget: 1000, BaseRevenue: 100, BudgetChangeRate: -0.2}
	res3, _ := simulateFinance(in3)
	if res3.RiskLevel != "高" || !strings.Contains(res3.Advice, "分期实施") {
		t.Fatalf("净影响 -20%% 应高风险: %+v", res3)
	}
	// 参数白名单
	if _, err := simulateFinance(SimInput{BaseBudget: 0}); err == nil {
		t.Fatal("基期预算 0 应报错")
	}
	if _, err := simulateFinance(SimInput{BaseBudget: 100, BudgetChangeRate: 1.5}); err == nil {
		t.Fatal("增减率超界应报错")
	}
}

// TestCompareSims 方案对比
func TestCompareSims(t *testing.T) {
	a := &SimResult{Net: -50, RiskLevel: "中"}
	b := &SimResult{Net: 20, RiskLevel: "低"}
	s := compareSims(a, b)
	if !strings.Contains(s, "方案 B 净影响优 +70") {
		t.Fatalf("对比摘要不符: %s", s)
	}
	same := compareSims(&SimResult{Net: 10, RiskLevel: "低"}, &SimResult{Net: 10, RiskLevel: "中"})
	if !strings.Contains(same, "相同") {
		t.Fatalf("持平口径不符: %s", same)
	}
}
