// Package service llmops 记忆域单测（窗口截断/上下文块格式/抽取解析）
package service

import (
	"strings"
	"testing"
)

func TestTruncateHistoryWindow(t *testing.T) {
	mk := func(i int) ChatMessage { return ChatMessage{Role: "user", Content: string(rune('a' + i))} }
	// 少于窗口原样
	msgs := []ChatMessage{mk(0), mk(1), mk(2)}
	if got := TruncateHistoryWindow(msgs, 5); len(got) != 3 {
		t.Fatalf("不足窗口应原样: %d", len(got))
	}
	// 超窗保留最近 N*2
	var many []ChatMessage
	for i := 0; i < 20; i++ {
		many = append(many, mk(i))
	}
	got := TruncateHistoryWindow(many, 3)
	if len(got) != 6 || got[len(got)-1].Content != "t" { // 第 20 个 = 'a'+19 = 't'
		t.Fatalf("应保留最近 6 条且末条最新: %d %q", len(got), got[len(got)-1].Content)
	}
	// 零值默认 5 轮
	if got := TruncateHistoryWindow(many, 0); len(got) != 10 {
		t.Fatalf("默认应 5 轮=10 条: %d", len(got))
	}
}

func TestBuildMemoryContext(t *testing.T) {
	if got := buildMemoryContext(nil); got != "" {
		t.Fatalf("空记忆应空串: %q", got)
	}
	hits := []FtHit{{Content: "用户偏好中文回答"}, {Content: "用户在运维团队"}}
	got := buildMemoryContext(hits)
	if !strings.HasPrefix(got, "[长期记忆]") || !strings.Contains(got, "1. 用户偏好中文回答") || !strings.Contains(got, "2. 用户在运维团队") {
		t.Fatalf("记忆块格式异常: %q", got)
	}
}

func TestParseMemoryExtract(t *testing.T) {
	// 有记忆
	fact, ok := ParseMemoryExtract(`{"memory":"用户名张三"}`)
	if !ok || fact != "用户名张三" {
		t.Fatalf("有记忆解析异常: %q %v", fact, ok)
	}
	// 围栏 + null
	if _, ok := ParseMemoryExtract("```json\n{\"memory\":null}\n```"); ok {
		t.Fatal("null 应视为无记忆")
	}
	// 空串记忆
	if _, ok := ParseMemoryExtract(`{"memory":"  "}`); ok {
		t.Fatal("空白记忆应视为无")
	}
	// 非 JSON
	if _, ok := ParseMemoryExtract("没有对象"); ok {
		t.Fatal("非 JSON 应无记忆")
	}
}
