// Package service exam 防作弊单测（事件白名单/编辑相似度/可疑对抽查）
package service

import "testing"

func TestValidBehaviorEvent(t *testing.T) {
	for _, ok := range BehaviorEvents {
		if !ValidBehaviorEvent(ok) {
			t.Fatalf("合法事件 %q 被拒绝", ok)
		}
	}
	for _, bad := range []string{"", "BLUR", "screenshot", "blur;"} {
		if ValidBehaviorEvent(bad) {
			t.Fatalf("非法事件 %q 被放行", bad)
		}
	}
}

func TestEditSimilarity(t *testing.T) {
	if got := editSimilarity("相同答案", "相同答案"); got != 1 {
		t.Fatalf("全同应 1: %v", got)
	}
	if got := editSimilarity("", "  "); got != 1 {
		t.Fatalf("双空(trim 后)应 1: %v", got)
	}
	if got := editSimilarity("abc", ""); got != 0 {
		t.Fatalf("一空一非空应 0: %v", got)
	}
	s1 := editSimilarity("SFT 是有监督微调", "SFT 是有监督微调方法")
	s2 := editSimilarity("SFT 是有监督微调", "今天天气不错")
	if s1 <= s2 {
		t.Fatalf("相似度排序异常: %v vs %v", s1, s2)
	}
}

func TestFindSuspiciousPairs(t *testing.T) {
	answers := []SheetAnswer{
		{SheetId: 1, Answer: "SFT 是有监督微调方法，LoRA 是低秩适配"},
		{SheetId: 2, Answer: "SFT 是有监督微调方法，LoRA 是低秩适配"}, // 抄袭
		{SheetId: 3, Answer: "完全不同的个人见解内容甲"},
		{SheetId: 4, Answer: "另一种独立的作答思路内容乙"},
	}
	pairs := findSuspiciousPairs(answers, 10)
	if len(pairs) != 1 {
		t.Fatalf("应仅 1 可疑对: %v", pairs)
	}
	p := pairs[0]
	if (p.SheetA != 1 || p.SheetB != 2) && (p.SheetA != 2 || p.SheetB != 1) {
		t.Fatalf("可疑对应为 1-2: %+v", p)
	}
	if p.Ratio < cheatSimThreshold {
		t.Fatalf("比率应过阈值: %v", p)
	}
	// 自身不比（同 SheetId 跳过）
	same := []SheetAnswer{{SheetId: 1, Answer: "x"}, {SheetId: 1, Answer: "x"}}
	if got := findSuspiciousPairs(same, 10); len(got) != 0 {
		t.Fatalf("同卷不应自比: %v", got)
	}
	// 全不同 → 空
	distinct := []SheetAnswer{{SheetId: 1, Answer: "甲"}, {SheetId: 2, Answer: "乙"}}
	if got := findSuspiciousPairs(distinct, 10); len(got) != 0 {
		t.Fatalf("无抄袭应空: %v", got)
	}
	// 少于两卷
	if got := findSuspiciousPairs(answers[:1], 10); len(got) != 0 {
		t.Fatal("单卷应空")
	}
	// sampleSize 截断
	many := []SheetAnswer{}
	for i := 0; i < 10; i++ {
		many = append(many, SheetAnswer{SheetId: uint(100 + i), Answer: "完全相同的集体模板答案"})
	}
	got := findSuspiciousPairs(many, 3)
	if len(got) != 3 {
		t.Fatalf("sampleSize 截断异常: %d", len(got))
	}
}
