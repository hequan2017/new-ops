package service

import "testing"

func TestEvaluateMetric(t *testing.T) {
	// 全部 >80 → 触发
	if !evaluateMetric([]float64{91, 85, 88}, ">", 80) {
		t.Fatal("连续超阈值应触发")
	}
	// 任一不满足 → 不触发
	if evaluateMetric([]float64{91, 75, 88}, ">", 80) {
		t.Fatal("存在未超阈值点不应触发")
	}
	// < 阈值（如磁盘剩余）
	if !evaluateMetric([]float64{5, 4}, "<", 10) {
		t.Fatal("< 阈值应触发")
	}
	// 空序列
	if evaluateMetric(nil, ">", 0) {
		t.Fatal("空序列不应触发")
	}
}
