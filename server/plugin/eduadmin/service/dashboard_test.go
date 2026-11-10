// Package service eduadmin：数据驾驶舱单测（纯函数，无 DB）
package service

import (
	"testing"
)

// TestBuildEduTrend 趋势补零
func TestBuildEduTrend(t *testing.T) {
	today := ""
	_ = today
	rows := []struct {
		Day string
		Cnt int64
	}{
		{Day: timeNowDash().Format("2006-01-02"), Cnt: 5},
	}
	trend := buildEduTrend(rows, 7)
	if len(trend) != 7 {
		t.Fatalf("应 7 点: %d", len(trend))
	}
	if trend[6].Count != 5 {
		t.Fatalf("今日应 5: %+v", trend[6])
	}
	for i := 0; i < 6; i++ {
		if trend[i].Count != 0 {
			t.Fatalf("缺口日应补零: %+v", trend[i])
		}
	}
	if trend := buildEduTrend(nil, 0); len(trend) != 7 {
		t.Fatalf("默认窗口 7: %d", len(trend))
	}
}
