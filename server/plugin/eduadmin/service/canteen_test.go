// Package service eduadmin：食堂优化单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// TestWeekdayMean 同星期日均人次
func TestWeekdayMean(t *testing.T) {
	records := []model.EduMealRecord{
		{Date: "2026-11-02", Stall: "面档", MealTime: "午餐", Count: 100}, // 周一
		{Date: "2026-11-09", Stall: "面档", MealTime: "午餐", Count: 120}, // 周一
		{Date: "2026-11-03", Stall: "面档", MealTime: "午餐", Count: 200}, // 周二
	}
	// 周一均值 (100+120)/2 = 110
	if v := weekdayMean(records, 1); v != 110 {
		t.Fatalf("周一均值应 110: %v", v)
	}
	if v := weekdayMean(records, 5); v != 0 {
		t.Fatalf("周五无数据应为 0: %v", v)
	}
}

// TestMealSuggestions 食材与菜品建议
func TestMealSuggestions(t *testing.T) {
	records := []model.EduMealRecord{
		{Date: "2026-11-02", Stall: "面档", MealTime: "午餐", Count: 300},
		{Date: "2026-11-02", Stall: "卤味档", MealTime: "午餐", Count: 50},
		{Date: "2026-11-02", Stall: "面档", MealTime: "早餐", Count: 80},
	}
	out := mealSuggestions(records, 200)
	joined := strings.Join(out, "|")
	if !strings.Contains(joined, "预计用餐 200 人次：按 230 人次备货") {
		t.Fatalf("备货建议应含 15%% 余量: %s", joined)
	}
	if !strings.Contains(joined, "【午餐】为用餐高峰") {
		t.Fatalf("午餐应为高峰: %s", joined)
	}
	if !strings.Contains(joined, "【卤味档】消费垫底") {
		t.Fatalf("卤味档应垫底: %s", joined)
	}
}
