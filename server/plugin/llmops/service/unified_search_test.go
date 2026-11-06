// Package service llmops 统一检索单测（模式白名单/日志截断）
package service

import "testing"

func TestValidSearchMode(t *testing.T) {
	for _, ok := range []string{"", "hybrid", "vector", "fulltext", "qa", "graph", "summary"} {
		if !ValidSearchMode(ok) {
			t.Fatalf("合法模式 %q 被拒绝", ok)
		}
	}
	for _, bad := range []string{"all", "HYBRID", "kg"} {
		if ValidSearchMode(bad) {
			t.Fatalf("非法模式 %q 被放行", bad)
		}
	}
}
