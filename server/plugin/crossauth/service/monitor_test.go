// Package service crossauth：五维监控单测（纯函数，无 DB）
package service

import (
	"testing"
)

// TestFiveDimHealthy 健康判定
func TestFiveDimHealthy(t *testing.T) {
	if !fiveDimHealthy(0, 0) {
		t.Fatal("零未处置应健康")
	}
	if fiveDimHealthy(1, 0) {
		t.Fatal("有未处置告警应不健康")
	}
	if fiveDimHealthy(0, 2) {
		t.Fatal("算力未处置应不健康")
	}
}
