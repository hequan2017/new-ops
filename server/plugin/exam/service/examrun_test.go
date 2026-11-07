// Package service exam 考试场次单测（状态机/时间窗）
package service

import (
	"testing"
	"time"
)

func TestCanTransitionExam(t *testing.T) {
	legal := []struct{ from, to string }{
		{"draft", "published"}, {"draft", "finished"},
		{"published", "ongoing"}, {"published", "draft"}, {"published", "finished"},
		{"ongoing", "finished"},
	}
	for _, c := range legal {
		if !CanTransitionExam(c.from, c.to) {
			t.Fatalf("合法流转 %s→%s 被拒绝", c.from, c.to)
		}
	}
	illegal := []struct{ from, to string }{
		{"draft", "ongoing"}, {"ongoing", "published"}, {"ongoing", "draft"},
		{"finished", "published"}, {"finished", "ongoing"},
	}
	for _, c := range illegal {
		if CanTransitionExam(c.from, c.to) {
			t.Fatalf("非法流转 %s→%s 被放行", c.from, c.to)
		}
	}
}

func TestValidateExamWindow(t *testing.T) {
	start := time.Date(2026, 11, 8, 9, 0, 0, 0, time.Local)
	end := time.Date(2026, 11, 8, 11, 0, 0, 0, time.Local)
	if err := ValidateExamWindow(start, end); err != nil {
		t.Fatalf("合法窗口应放行: %v", err)
	}
	if err := ValidateExamWindow(end, start); err == nil {
		t.Fatal("start>end 应拒绝")
	}
	if err := ValidateExamWindow(start, start); err == nil {
		t.Fatal("相等应拒绝")
	}
}
