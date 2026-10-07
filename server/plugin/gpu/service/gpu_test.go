package service

import "testing"

// TestCanAllocate 配额纯逻辑（防超卖判定核心）
func TestCanAllocate(t *testing.T) {
	if !canAllocate(8, 5, 3) {
		t.Fatal("余 3 需 3 应可分配")
	}
	if canAllocate(8, 6, 3) {
		t.Fatal("余 2 需 3 应拒绝")
	}
	if canAllocate(8, 0, 0) {
		t.Fatal("want=0 应拒绝")
	}
	if canAllocate(8, 9, 1) {
		t.Fatal("used>total 异常态也应拒绝")
	}
	if !canAllocate(4, 0, 4) {
		t.Fatal("恰好用满应可分配")
	}
}
