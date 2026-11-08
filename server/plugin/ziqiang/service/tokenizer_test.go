// Package service ziqiang：法律分词与抄袭识别单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestLegalSegment 复合术语词典命中
func TestLegalSegment(t *testing.T) {
	tokens := LegalSegment("本案涉及在先使用权与等同侵权的争议，被告主张禁止反悔。")
	words := map[string]bool{}
	for _, tk := range tokens {
		words[tk.Word] = tk.IsDict
	}
	for _, term := range []string{"在先使用权", "等同侵权", "禁止反悔"} {
		if !words[term] {
			t.Fatalf("应切出术语「%s」: %v", term, words)
		}
	}
	// 拉丁串成词、标点跳过
	tokens2 := LegalSegment("依据 CN202610123456.7 号专利，原告主张损害赔偿。")
	joined := ""
	for _, tk := range tokens2 {
		joined += tk.Word + "|"
	}
	if !strings.Contains(joined, "CN202610123456.7") || !strings.Contains(joined, "损害赔偿") {
		t.Fatalf("拉丁串与术语不符: %s", joined)
	}
	if strings.Contains(joined, "，") || strings.Contains(joined, "。") {
		t.Fatalf("标点不应成词: %s", joined)
	}
}

// TestSlidingPlagiarism 嵌入抄袭命中、无关零命中
func TestSlidingPlagiarism(t *testing.T) {
	src := "专利权的保护范围以权利要求书的内容为准，说明书及附图可以用于解释权利要求，等同侵权判定适用禁止反悔原则，现有技术抗辩成立时不构成侵权。"
	// 嵌入式抄袭：rune 安全切片取源文长片段混入（越界按实际长度截）
	srcRunes := []rune(src)
	end := 170
	if end > len(srcRunes) {
		end = len(srcRunes)
	}
	fragment := string(srcRunes[10:end])
	susp := "本文研究了新型储能装置。据了解，" + fragment + "。此外作者提出了自己的观点，与上述内容无关的补充论述，量子计算与电池材料科学交叉领域还有诸多开放问题值得进一步研究探讨。"
	hits, err := SlidingPlagiarism(src, susp, 30, 0.8)
	if err != nil {
		t.Fatalf("检测不应报错: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("嵌入式抄袭应命中")
	}
	if hits[0].Containment < 0.8 {
		t.Fatalf("包含率应达标: %v", hits[0].Containment)
	}
	// 完全无关文本
	hits2, err := SlidingPlagiarism(src, "今天天气很好，适合去公园散步和野餐，孩子们在草地上放风筝，非常惬意的一个下午。", 40, 0.8)
	if err != nil || len(hits2) != 0 {
		t.Fatalf("无关文本应零命中: %v %v", hits2, err)
	}
}

// TestSlidingPlagiarismGuard 参数守卫
func TestSlidingPlagiarismGuard(t *testing.T) {
	if _, err := SlidingPlagiarism("src", "susp", 5, 0.8); err == nil {
		t.Fatal("窗口过小应报错")
	}
	if _, err := SlidingPlagiarism("src", "susp", 40, 0.1); err == nil {
		t.Fatal("阈值过低应报错")
	}
}
