// Package service llmops 知识库单测（类型白名单）
package service

import "testing"

func TestValidKnowledgeType(t *testing.T) {
	for _, ok := range []string{"vector", "fulltext", "graph", "qa"} {
		if !ValidKnowledgeType(ok) {
			t.Fatalf("合法类型 %q 被拒绝", ok)
		}
	}
	for _, bad := range []string{"", "Vector", "kg", "text"} {
		if ValidKnowledgeType(bad) {
			t.Fatalf("非法类型 %q 被放行", bad)
		}
	}
}
