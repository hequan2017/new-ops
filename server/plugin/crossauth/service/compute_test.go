// Package service crossauth：算力防护单测（纯函数，无 DB）
package service

import (
	"testing"
)

// TestComputeDecision 阈值三档
func TestComputeDecision(t *testing.T) {
	cases := map[float64]string{
		500:  "放行",
		800:  "放行",   // 恰 800 未超软限
		800.1: "降优先级",
		1000: "降优先级", // 恰 1000 未超硬限
		1000.1: "限流",
		2000: "限流",
	}
	for hours, want := range cases {
		if got := computeDecision(hours); got != want {
			t.Fatalf("%.1f 核时应 %s 得 %s", hours, want, got)
		}
	}
}
