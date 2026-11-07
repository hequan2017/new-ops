// Package service llmops 监控域单测（时延四桶分布边界）
package service

import "testing"

func TestLatencyBuckets(t *testing.T) {
	got := LatencyBuckets([]int64{100, 499, 500, 1999, 2000, 9999, 10000, 50000})
	want := [4]int64{2, 2, 2, 2}
	if got != want {
		t.Fatalf("四桶边界异常: %v want %v", got, want)
	}
	if LatencyBuckets(nil) != [4]int64{} {
		t.Fatal("空输入应全零")
	}
	// 单桶
	if got := LatencyBuckets([]int64{300}); got != [4]int64{1, 0, 0, 0} {
		t.Fatalf("单样本: %v", got)
	}
}
