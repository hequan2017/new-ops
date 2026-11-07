// Package service exam 终端接入单测（可考场次过滤）
package service

import (
	"testing"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/exam/model"
)

func TestFilterOngoingExams(t *testing.T) {
	now := time.Date(2026, 11, 8, 10, 0, 0, 0, time.Local)
	mk := func(id uint, status string, start, end time.Time) model.ExamExam {
		e := model.ExamExam{Name: "e", Status: status, StartTime: start, EndTime: end, DurationMin: 60}
		e.ID = id
		return e
	}
	inWindow := func(offset time.Duration) (time.Time, time.Time) {
		return now.Add(offset), now.Add(offset + 2*time.Hour)
	}
	s, e := inWindow(-30 * time.Minute)
	exams := []model.ExamExam{
		mk(1, "ongoing", s, e),                        // ✓ ongoing 且窗内
		mk(2, "published", s, e),                      // ✗ 未开始流转
		mk(3, "ongoing", now.Add(time.Hour), now.Add(3*time.Hour)), // ✗ 未到开考
		mk(4, "ongoing", now.Add(-3*time.Hour), now.Add(-time.Hour)), // ✗ 已过截止
		mk(5, "finished", s, e),                       // ✗ 已结束
	}
	got := filterOngoingExams(exams, now)
	if len(got) != 1 || got[0].ExamId != 1 {
		t.Fatalf("应仅窗内 ongoing: %+v", got)
	}
	if got[0].DurationMin != 60 || got[0].Name != "e" {
		t.Fatalf("字段映射异常: %+v", got[0])
	}
	// 空列表
	if got := filterOngoingExams(nil, now); len(got) != 0 {
		t.Fatal("空输入应空")
	}
}
