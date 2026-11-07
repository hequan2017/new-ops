// Package service exam 答题链路单测（脱敏/乐观锁/时间窗纯函数）
package service

import (
	"strings"
	"testing"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/exam/model"
)

func TestStripAnswers(t *testing.T) {
	qs := []model.ExamQuestion{
		{Stem: "题一", Answer: "秘密", OptionsJSON: `["A","B"]`},
		{Stem: "题二", Answer: "另一个秘密"},
	}
	got := stripAnswers(qs)
	for _, q := range got {
		if q.Answer != "" {
			t.Fatalf("答案未剥除: %q", q.Answer)
		}
	}
	if got[0].OptionsJSON != `["A","B"]` {
		t.Fatalf("选项应保留: %q", got[0].OptionsJSON)
	}
	if got[0].Stem != "题一" {
		t.Fatal("题干应保留")
	}
	// 原 slice 不被污染
	if qs[0].Answer != "秘密" {
		t.Fatal("原数据不应被修改")
	}
}

func TestCheckVersion(t *testing.T) {
	if err := checkVersion(3, 3); err != nil {
		t.Fatalf("同版本应放行: %v", err)
	}
	if err := checkVersion(3, -1); err != nil {
		t.Fatalf("强制模式应放行: %v", err)
	}
	err := checkVersion(3, 2)
	if err == nil || !strings.Contains(err.Error(), "2413") {
		t.Fatalf("版本冲突应报 2413: %v", err)
	}
}

func TestWithinExamWindow(t *testing.T) {
	start := time.Date(2026, 11, 8, 9, 0, 0, 0, time.Local)
	end := time.Date(2026, 11, 8, 11, 0, 0, 0, time.Local)
	now := time.Date(2026, 11, 8, 10, 0, 0, 0, time.Local)
	if err := withinExamWindow(start, end, now); err != nil {
		t.Fatalf("窗内应放行: %v", err)
	}
	if err := withinExamWindow(start, end, start.Add(-time.Minute)); err == nil {
		t.Fatal("未开始应拒绝")
	}
	if err := withinExamWindow(start, end, end.Add(time.Minute)); err == nil {
		t.Fatal("已结束应拒绝")
	}
}
