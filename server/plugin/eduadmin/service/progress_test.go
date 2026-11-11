// Package service eduadmin：工单督办单测（纯函数，无 DB）
package service

import (
	"testing"
	"time"
)

// TestUrgeLevel 督办三档
func TestUrgeLevel(t *testing.T) {
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.Local)
	cases := []struct {
		deadline time.Time
		want     string
	}{
		{now.Add(-1 * time.Hour), "超时"},
		{now.Add(3 * time.Hour), "临近"},  // <6h
		{now.Add(5*time.Hour + 59*time.Minute), "临近"},
		{now.Add(6 * time.Hour), ""},    // 恰 6h 不提醒
		{now.Add(24 * time.Hour), ""},
	}
	for _, c := range cases {
		if got := urgeLevel(now, c.deadline); got != c.want {
			t.Fatalf("deadline=%v 应 %s 得 %s", c.deadline, c.want, got)
		}
	}
}

// TestTicketDeadlineTime 时限推算
func TestTicketDeadlineTime(t *testing.T) {
	created := time.Date(2026, 11, 10, 9, 0, 0, 0, time.Local)
	deadline, err := ticketDeadlineTime(created, "简单24h")
	if err != nil || !deadline.Equal(created.Add(24*time.Hour)) {
		t.Fatalf("简单24h 应 +24h: %v %v", deadline, err)
	}
	deadline, _ = ticketDeadlineTime(created, "复杂7工作日")
	if deadline.Sub(created) != 168*time.Hour {
		t.Fatalf("复杂 7 工作日应 +168h: %v", deadline.Sub(created))
	}
	if _, err := ticketDeadlineTime(created, "即时"); err == nil {
		t.Fatal("时限类别白名单外应报错")
	}
}
