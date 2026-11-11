// Package service crossauth：应急演练单测（纯函数，无 DB）
package service

import (
	"testing"
)

// TestDrillTargetMet 年度 2 次口径达成
func TestDrillTargetMet(t *testing.T) {
	if drillTargetMet(1, 2) {
		t.Fatal("1 次未达 2 次口径")
	}
	if !drillTargetMet(2, 2) {
		t.Fatal("2 次应达成")
	}
	if !drillTargetMet(3, 0) {
		t.Fatal("目标缺省 2，3 次应达成")
	}
}
