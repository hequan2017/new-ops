// Package service crossauth：联合运维单测（纯函数，无 DB）
package service

import (
	"testing"
)

// TestCaOpsNext 工单状态机
func TestCaOpsNext(t *testing.T) {
	stages := map[string]string{"提交": "分配", "分配": "处理", "处理": "闭环"}
	for from, to := range stages {
		got, err := caOpsNext(from)
		if err != nil || got != to {
			t.Fatalf("%s 应推进 %s: %s %v", from, to, got, err)
		}
	}
	if _, err := caOpsNext("闭环"); err == nil {
		t.Fatal("终态应不可推进")
	}
}
