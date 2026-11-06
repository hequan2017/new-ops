// Package service llmops 问答库单测（编辑距离/相似度/排序纯函数）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

func TestEditDistance(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "abc", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"kitten", "sitting", 3},
		{"白泽平台", "白泽运维平台", 2},
		{"flaw", "lawn", 2},
	}
	for _, c := range cases {
		if got := EditDistance([]rune(c.a), []rune(c.b)); got != c.want {
			t.Fatalf("EditDistance(%q,%q)=%d want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestQaSimilarity(t *testing.T) {
	// 归一键语义：连续空白折叠为单空格（A05 契约），双空格差异归一后一致
	if got := QaSimilarity("What  is  Baize", "what is baize"); got != 1 {
		t.Fatalf("归一后一致应得 1: %v", got)
	}
	if got := QaSimilarity("白泽是什么平台", "白泽是什么平台"); got != 1 {
		t.Fatalf("完全一致应得 1: %v", got)
	}
	s1 := QaSimilarity("什么是白泽", "什么是白泽平台")
	s2 := QaSimilarity("什么是白泽", "今天天气怎么样")
	if s1 <= s2 || s1 <= 0 || s2 < 0 {
		t.Fatalf("相似度排序异常: %v vs %v", s1, s2)
	}
	if s1 > 1 || s2 > 1 {
		t.Fatalf("相似度应 ≤1: %v %v", s1, s2)
	}
}

// TestRankQaPairs 精确优先、阈值过滤、排序稳定
func TestRankQaPairs(t *testing.T) {
	mk := func(id uint, q, a string) model.LlmopsQaPair {
		p := model.LlmopsQaPair{Question: q, Answer: a}
		p.ID = id
		return p
	}
	pairs := []model.LlmopsQaPair{
		mk(1, "什么是白泽", "统一运维平台"),
		mk(2, "白泽是什么", "另一个答案"),
		mk(3, "白泽是什么平台", "相近问题"),
		mk(4, "完全无关的内容呢", "无关答案"),
		mk(5, "什么是白泽", "重复问题"),
	}
	hits := rankQaPairs("什么是白泽", pairs, 0.1)
	// 精确命中(1,5)优先且按 pairId 升序；2(0.2)/3(0.286) 过 0.1 阈值；4(≈0) 被过滤
	if len(hits) != 4 {
		t.Fatalf("应 4 命中(4 被过滤): %+v", hits)
	}
	if !hits[0].Exact || hits[0].PairId != 1 || !hits[1].Exact || hits[1].PairId != 5 {
		t.Fatalf("精确命中应最优先按 ID 升序: %+v", hits[:2])
	}
	if hits[2].Exact || hits[3].Exact {
		t.Fatalf("非精确不应标 Exact: %+v", hits[2:])
	}
	if hits[2].Score < hits[3].Score {
		t.Fatalf("相似度应降序: %+v", hits[2:])
	}
	// 高阈值只剩精确
	strict := rankQaPairs("什么是白泽", pairs, 0.99)
	if len(strict) != 2 || !strict[0].Exact {
		t.Fatalf("高阈值应仅剩精确: %+v", strict)
	}
}
