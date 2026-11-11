// Package service eduadmin：反馈闭环单测（纯函数，无 DB）
package service

import (
	"testing"
)

// TestReviewOutcome 评价结果
func TestReviewOutcome(t *testing.T) {
	if next, err := reviewOutcome("好评"); err != nil || next != "评价" {
		t.Fatalf("好评应到评价态: %s %v", next, err)
	}
	if next, _ := reviewOutcome("差评"); next != "办理" {
		t.Fatalf("差评应回流办理: %s", next)
	}
	if _, err := reviewOutcome("中评"); err == nil {
		t.Fatal("白名单外应报错")
	}
}
