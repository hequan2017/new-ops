// Package service llmops simhash 单测（指纹确定性/海明距离/重复判定语义）
package service

import "testing"

func TestSimhash64(t *testing.T) {
	// 确定性
	a1 := Simhash64("白泽统一运维平台的文档内容")
	a2 := Simhash64("白泽统一运维平台的文档内容")
	if a1 != a2 {
		t.Fatal("同文本指纹应恒定")
	}
	// 相似文本指纹应接近（海明距离小）
	b := Simhash64("白泽统一运维平台的文档内容简介")
	if HammingDistance(a1, b) > 10 {
		t.Fatalf("相似文本指纹应接近: %d", HammingDistance(a1, b))
	}
	// 无关文本距离应大
	c := Simhash64("今天晚上吃火锅还是烧烤呢")
	if HammingDistance(a1, c) < 10 {
		t.Fatalf("无关文本距离应大: %d", HammingDistance(a1, c))
	}
	// 空文本不 panic
	_ = Simhash64("")
}

func TestHammingDistance(t *testing.T) {
	if got := HammingDistance(0, 0); got != 0 {
		t.Fatalf("同值应 0: %d", got)
	}
	if got := HammingDistance(0b1010, 0b0101); got != 4 {
		t.Fatalf("全异应 4: %d", got)
	}
	if got := HammingDistance(0b1100, 0b1000); got != 1 {
		t.Fatalf("单比特差应 1: %d", got)
	}
}

// TestDetectDocDuplicate_Threshold 阈值语义（≤3 判重）纯逻辑验证
func TestDetectDocDuplicate_Threshold(t *testing.T) {
	// 相同文本 dist=0 → 判重
	if d := HammingDistance(Simhash64("重复检测样本文本ABC"), Simhash64("重复检测样本文本ABC")); d > simhashHammingThreshold {
		t.Fatalf("同文本应在阈值内: %d", d)
	}
}
