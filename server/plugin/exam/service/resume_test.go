// Package service exam 断线续考单测（剩余时长双口径截断）
package service

import (
	"testing"
	"time"
)

func TestRemainingSeconds(t *testing.T) {
	now := time.Date(2026, 11, 8, 10, 0, 0, 0, time.Local)
	examEnd := now.Add(2 * time.Hour)
	started := now.Add(-30 * time.Minute)

	// 单场时长更短：duration 60 分钟，已用 30 → 剩 1800s
	if got := remainingSeconds(examEnd, started, 60, now); got != 1800 {
		t.Fatalf("单场口径应剩 1800: %d", got)
	}
	// 考试截止更短：duration 300 分钟 → 剩 2h=7200s
	if got := remainingSeconds(examEnd, started, 300, now); got != 7200 {
		t.Fatalf("截止口径应剩 7200: %d", got)
	}
	// 双超时 → 0
	if got := remainingSeconds(examEnd, started.Add(-time.Hour), 60, now); got != 0 {
		t.Fatalf("超时应 0: %d", got)
	}
	// duration 0（未配置）→ 仅按截止
	if got := remainingSeconds(examEnd, started, 0, now); got != 7200 {
		t.Fatalf("未配置时长应按截止: %d", got)
	}
	// startedAt 零值 → 仅按截止
	if got := remainingSeconds(examEnd, time.Time{}, 60, now); got != 7200 {
		t.Fatalf("零开始时间应按截止: %d", got)
	}
	// 已过截止 → 0
	if got := remainingSeconds(now.Add(-time.Minute), started, 60, now); got != 0 {
		t.Fatalf("过截止应 0: %d", got)
	}
}
