// Package service llmops 看板单测（趋势补零纯函数）
package service

import "testing"

func TestBuildTrend(t *testing.T) {
	// 缺口补零：仅中间一天有数据
	rows := []struct {
		Day  string
		Cnt  int64
	}{
		{Day: daysAgo(3), Cnt: 42},
	}
	got := buildTrend(rows, 7)
	if len(got) != 7 {
		t.Fatalf("应 7 点: %d", len(got))
	}
	// 第 4 点（3 天前）有值，其余 0
	mid := got[3]
	if mid.Date != daysAgo(3) || mid.Calls != 42 {
		t.Fatalf("有值日错位: %+v", mid)
	}
	zeroCount := 0
	for _, p := range got {
		if p.Calls == 0 {
			zeroCount++
		}
	}
	if zeroCount != 6 {
		t.Fatalf("缺口应补零: %d", zeroCount)
	}
	// 首尾日期有序
	if got[0].Date >= got[6].Date {
		t.Fatalf("日期应升序: %s ~ %s", got[0].Date, got[6].Date)
	}
	// 空输入全零
	got2 := buildTrend(nil, 7)
	if len(got2) != 7 {
		t.Fatalf("空输入应仍 7 点: %d", len(got2))
	}
	// days 夹取
	if got3 := buildTrend(nil, 0); len(got3) != 7 {
		t.Fatalf("days=0 应默认 7: %d", len(got3))
	}
}

func daysAgo(n int) string {
	return timeNowDash().AddDate(0, 0, -n).Format("2006-01-02")
}
