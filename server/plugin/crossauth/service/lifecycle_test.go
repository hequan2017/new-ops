// Package service crossauth：权限生命周期单测（纯函数，无 DB）
package service

import (
	"testing"
	"time"
)

// TestExpiryAction 三档决策
func TestExpiryAction(t *testing.T) {
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.Local)
	cases := []struct {
		expiredAt time.Time
		want      string
	}{
		{now.AddDate(0, 0, 3), "无"},      // 未来
		{now.AddDate(0, 0, 0), "无"},      // 当日未过
		{now.AddDate(0, 0, -2), "提醒"},    // 过期 2 天（剩余 ≤7 天窗口）
		{now.AddDate(0, 0, -7), "提醒"},    // 恰 7 天
		{now.AddDate(0, 0, -8), "回收"},    // 超 7 天
		{now.AddDate(0, -3, 0), "回收"},    // 3 个月前
	}
	for _, c := range cases {
		if got := expiryAction(now, c.expiredAt); got != c.want {
			t.Fatalf("到期 %v 应 %s 得 %s", c.expiredAt, c.want, got)
		}
	}
}

// TestDaysLeft 剩余天数
func TestDaysLeft(t *testing.T) {
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.Local)
	if d := daysLeft(now, now.Add(36*time.Hour)); d != 1 {
		t.Fatalf("36h 应 1 天: %d", d)
	}
	if d := daysLeft(now, now.Add(-24*time.Hour)); d != 0 {
		t.Fatalf("过期应 0: %d", d)
	}
}
