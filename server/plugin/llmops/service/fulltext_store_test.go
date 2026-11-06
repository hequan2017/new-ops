// Package service llmops 全文库单测（bigram 切分/BM25 打分语义）
package service

import "testing"

func TestBigrams(t *testing.T) {
	// CJK bigram
	got := Bigrams("白泽平台")
	if len(got) != 3 || got[0] != "白泽" || got[1] != "泽平" || got[2] != "平台" {
		t.Fatalf("CJK bigram 异常: %v", got)
	}
	// ASCII 词元（小写）
	got = Bigrams("Hello World")
	if len(got) != 2 || got[0] != "hello" || got[1] != "world" {
		t.Fatalf("ASCII 词元异常: %v", got)
	}
	// 混合：CJK 与 ASCII 连续无分隔 → 整 token bigram（跨脚本界一致口径）
	got = Bigrams("大模型LLM，检索！")
	if len(got) != 6 || got[0] != "大模" || got[1] != "模型" || got[2] != "型l" || got[3] != "ll" || got[4] != "lm" || got[5] != "检索" {
		t.Fatalf("混合切分异常: %v", got)
	}
	// 有标点分隔则 ASCII 独立词元
	got = Bigrams("大模型 LLM，检索")
	if len(got) != 4 || got[0] != "大模" || got[1] != "模型" || got[2] != "llm" || got[3] != "检索" {
		t.Fatalf("空格分隔切分异常: %v", got)
	}
	// 单字 CJK 保留
	got = Bigrams("龙")
	if len(got) != 1 || got[0] != "龙" {
		t.Fatalf("单字应保留: %v", got)
	}
	// 空串/纯标点
	if len(Bigrams("，。！")) != 0 {
		t.Fatal("纯标点应为空")
	}
}

// TestBM25Score 打分语义：词频↑得分↑；文档变长稀释；未命中词零贡献
func TestBM25Score(t *testing.T) {
	df := map[string]int{"白泽": 2, "平台": 3}
	docFreq := map[string]int{"白泽": 2, "平台": 1}
	base := BM25Score([]string{"白泽", "平台"}, docFreq, 50, 50, df, 10)
	if base <= 0 {
		t.Fatalf("命中应得分: %v", base)
	}
	// 词频翻倍 → 得分升但次线性
	docFreq2 := map[string]int{"白泽": 4, "平台": 1}
	higher := BM25Score([]string{"白泽", "平台"}, docFreq2, 50, 50, df, 10)
	if higher <= base {
		t.Fatalf("词频增应得分增: %v -> %v", base, higher)
	}
	// 文档更长（同词频）→ 得分降
	longer := BM25Score([]string{"白泽", "平台"}, docFreq, 200, 50, df, 10)
	if longer >= base {
		t.Fatalf("长文档应稀释得分: %v -> %v", base, longer)
	}
	// 未命中词
	if got := BM25Score([]string{"不存在"}, docFreq, 50, 50, df, 10); got != 0 {
		t.Fatalf("未命中应为 0: %v", got)
	}
	// 空文档
	if got := BM25Score([]string{"白泽"}, map[string]int{}, 0, 50, df, 10); got != 0 {
		t.Fatalf("空文档应为 0: %v", got)
	}
	// 查询词去重（同词重复不计双份）
	dup := BM25Score([]string{"白泽", "白泽", "平台"}, docFreq, 50, 50, df, 10)
	if dup != base {
		t.Fatalf("查询词应去重: %v vs %v", dup, base)
	}
}
