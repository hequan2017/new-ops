// Package service eduadmin：AI-12345 工单单测（纯函数，无 DB）
package service

import (
	"testing"
)

// TestTicketDeadlineHours 时限类别
func TestTicketDeadlineHours(t *testing.T) {
	cases := map[string]int{"简单24h": 24, "一般3工作日": 72, "复杂7工作日": 168}
	for class, want := range cases {
		got, err := TicketDeadlineHours(class)
		if err != nil || got != want {
			t.Fatalf("%s 应 %dh: %d %v", class, want, got, err)
		}
	}
	if _, err := TicketDeadlineHours("即时"); err == nil {
		t.Fatal("白名单外应报错")
	}
}

// TestNextTicketStatus 状态机线性流转
func TestNextTicketStatus(t *testing.T) {
	want := map[string]string{"受理": "分派", "分派": "办理", "办理": "办结", "办结": "评价", "评价": "归档"}
	for from, to := range want {
		got, err := nextTicketStatus(from)
		if err != nil || got != to {
			t.Fatalf("%s 应流转 %s: %s %v", from, to, got, err)
		}
	}
	if _, err := nextTicketStatus("归档"); err == nil {
		t.Fatal("终态应报错")
	}
	if _, err := nextTicketStatus("未知"); err == nil {
		t.Fatal("未知状态应报错")
	}
}
