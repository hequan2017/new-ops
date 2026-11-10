// Package service eduadmin：智能采购单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// TestAggregateMonthly 按月聚合（同年同月合并、升序）
func TestAggregateMonthly(t *testing.T) {
	list := []model.EduPurchase{
		{Year: 2026, Month: 3, Item: "耗材", Qty: 10},
		{Year: 2026, Month: 1, Item: "耗材", Qty: 8},
		{Year: 2026, Month: 3, Item: "耗材", Qty: 5},
		{Year: 2025, Month: 12, Item: "耗材", Qty: 12},
	}
	months, qtys := aggregateMonthly(list)
	if len(qtys) != 3 {
		t.Fatalf("应聚合为 3 期: %v", qtys)
	}
	if qtys[0] != 12 || qtys[1] != 8 || qtys[2] != 15 {
		t.Fatalf("聚合量应为 12/8/15: %v", qtys)
	}
	if months[0] != 2025*12+12 {
		t.Fatalf("应升序: %v", months)
	}
}

// TestSuggestPurchase 建议生成（量/时点/预算）
func TestSuggestPurchase(t *testing.T) {
	list := []model.EduPurchase{
		{Year: 2025, Month: 9, Item: "笔记本", Qty: 40, UnitPrice: 6000},
		{Year: 2026, Month: 9, Item: "笔记本", Qty: 50, UnitPrice: 6000},
		{Year: 2026, Month: 10, Item: "笔记本", Qty: 30, UnitPrice: 5800},
		{Year: 2026, Month: 11, Item: "笔记本", Qty: 45, UnitPrice: 5800},
	}
	out, err := suggestPurchase("笔记本", list)
	if err != nil {
		t.Fatalf("建议失败: %v", err)
	}
	// MA3 = (50+30+45)/3 = 41.67
	if out.ForecastQty != 41.67 {
		t.Fatalf("预测量应 41.67: %v", out.ForecastQty)
	}
	// 峰值月 9（出现最多）→ 提前 1 月=8
	if out.OrderMonth != 8 {
		t.Fatalf("应 8 月下单: %d", out.OrderMonth)
	}
	if out.UnitPrice != 5800 {
		t.Fatalf("最近单价应 5800: %v", out.UnitPrice)
	}
	if out.Budget != 41.67*5800 {
		t.Fatalf("预算=量×价: %v", out.Budget)
	}
	if !strings.Contains(out.Summary, "8 月下单") {
		t.Fatalf("摘要不符: %s", out.Summary)
	}
	if _, err := suggestPurchase("新品", nil); err == nil {
		t.Fatal("无历史应报错")
	}
}
