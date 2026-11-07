// Package service exam 考生管理单测（券码生成/名单去重纯函数）
package service

import (
	"math/rand"
	"strings"
	"testing"
)

func TestGenerateTicketCodes(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	codes := GenerateTicketCodes(1000, rng)
	if len(codes) != 1000 {
		t.Fatalf("数量异常: %d", len(codes))
	}
	seen := map[string]bool{}
	for _, c := range codes {
		if len(c) != ticketCodeLen {
			t.Fatalf("长度异常: %q", c)
		}
		for _, ch := range c {
			if !strings.ContainsRune(ticketCodeAlphabet, ch) {
				t.Fatalf("字符集异常: %q", c)
			}
		}
		if seen[c] {
			t.Fatalf("千级样本不应碰撞: %q", c)
		}
		seen[c] = true
	}
}

func TestFilterDupCandidates(t *testing.T) {
	existing := map[string]bool{"old-user": true}
	valid, skipped := filterDupCandidates(existing, []string{"张三", "李四", "张三", "  ", "old-user"})
	// 张三/李四有效；张三重名、空串、old-user（userKey 集合仅作占位口径）→ 3 跳过
	if len(valid) != 2 || skipped != 3 {
		t.Fatalf("valid=%v skipped=%d want 2/3", valid, skipped)
	}
	if valid[0] != "张三" || valid[1] != "李四" {
		t.Fatalf("顺序异常: %v", valid)
	}
}
