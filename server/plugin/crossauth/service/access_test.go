// Package service crossauth：访问校验单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestAccessVerdict 双门判定四象限
func TestAccessVerdict(t *testing.T) {
	allow, reason := accessVerdict(true, true, "矩阵通过", "互认通过")
	if !allow || !strings.Contains(reason, "均放行") {
		t.Fatalf("双过应放行: %s", reason)
	}
	allow, reason = accessVerdict(false, true, "无命中", "互认通过")
	if allow || !strings.Contains(reason, "矩阵拒绝") {
		t.Fatalf("矩阵拒应阻断: %s", reason)
	}
	allow, reason = accessVerdict(true, false, "矩阵通过", "无匹配规则")
	if allow || !strings.Contains(reason, "互认拒绝") {
		t.Fatalf("互认拒应阻断: %s", reason)
	}
	allow, reason = accessVerdict(false, false, "m", "t")
	if allow || !strings.Contains(reason, "矩阵拒绝") || !strings.Contains(reason, "互认拒绝") {
		t.Fatalf("双拒应合并原因: %s", reason)
	}
}
