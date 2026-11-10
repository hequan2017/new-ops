// Package service eduadmin：智能预算单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// TestBudgetTrend 同比增长率
func TestBudgetTrend(t *testing.T) {
	list := []model.EduBudget{
		{Year: 2024, Dept: "计算机学院", Amount: 100},
		{Year: 2023, Dept: "计算机学院", Amount: 80},
		{Year: 2025, Dept: "计算机学院", Amount: 110},
	}
	trend := budgetTrend(list)
	if len(trend) != 3 || trend[0].Year != 2023 {
		t.Fatalf("应按年升序: %+v", trend)
	}
	if trend[1].Growth != 0.25 || trend[2].Growth != 0.1 {
		t.Fatalf("同比应 0.25/0.1: %+v", trend)
	}
	if trend[0].Growth != 0 {
		t.Fatalf("首年增长率应为 0: %+v", trend[0])
	}
}

// TestSuggestAllocation 占比缩放+下限重归一
func TestSuggestAllocation(t *testing.T) {
	history := []deptAmount{
		{Dept: "教学A", Amount: 600},
		{Dept: "教学B", Amount: 300},
		{Dept: "行政C", Amount: 100},
	}
	out := suggestAllocation(history, 1000)
	if len(out) != 3 {
		t.Fatalf("应 3 部门: %+v", out)
	}
	if out[0].Ratio != 0.6 || out[0].Amount != 600 {
		t.Fatalf("教学A 应保持 60%%: %+v", out[0])
	}
	if out[2].Ratio != 0.1 || out[2].Amount != 100 {
		t.Fatalf("行政C 下限 10%%: %+v", out[2])
	}
	var sumRatio float64
	for _, o := range out {
		sumRatio += o.Ratio
	}
	if sumRatio < 0.999 || sumRatio > 1.001 {
		t.Fatalf("占比应归一(±0.001): %v", sumRatio)
	}
	if suggestAllocation(nil, 1000) != nil {
		t.Fatal("空历史应返回空")
	}
}

// TestComparePlans 方案对比
func TestComparePlans(t *testing.T) {
	a := []deptAmount{{Dept: "教学", Amount: 700}, {Dept: "行政", Amount: 300}}
	b := []deptAmount{{Dept: "教学", Amount: 600}, {Dept: "行政", Amount: 500}}
	diff := comparePlans(a, b)
	if diff.TotalA != 1000 || diff.TotalB != 1100 || diff.Delta != 100 {
		t.Fatalf("总额差额应 100: %+v", diff)
	}
	if !strings.Contains(diff.Summary, "教学: 700→600 (-100)") || !strings.Contains(diff.Summary, "更倾斜 行政") {
		t.Fatalf("对比摘要不符: %s", diff.Summary)
	}
}
