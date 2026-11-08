// Package service fanhua：分词器单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestSegmentDict 词典词优先最长匹配
func TestSegmentDict(t *testing.T) {
	tokens := Segment("侬好，今朝到弄堂口看滑稽戏")
	var words []string
	for _, tk := range tokens {
		words = append(words, tk.Word)
	}
	joined := strings.Join(words, "|")
	for _, want := range []string{"侬好", "今朝", "弄堂", "滑稽戏"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("应切出词典词 %s: %s", want, joined)
		}
	}
	if len(words) > 0 && words[0] != "侬好" {
		t.Fatalf("最长匹配应优先切出「侬好」而非「侬」: %s", words[0])
	}
	// 单字词典词与多字词典词混合
	tokens = Segment("迭个小囡老嗲")
	hasDuo, hasDa := false, false
	for _, tk := range tokens {
		if tk.Word == "迭个" && tk.IsDict {
			hasDuo = true
		}
		if tk.Word == "嗲" && tk.IsDict {
			hasDa = true
		}
	}
	if !hasDuo || !hasDa {
		t.Fatalf("「迭个」「嗲」应词典命中: %+v", tokens)
	}
}

// TestSegmentOOVBigram 未登录汉字 bigram 兜底
func TestSegmentOOVBigram(t *testing.T) {
	// 「龘龘龘」均不在词典 → bigram 两两成词，奇数尾单字
	tokens := Segment("龘龘龘")
	if len(tokens) != 2 {
		t.Fatalf("3 个未登录字应成 2 token(2+1): %+v", tokens)
	}
	if tokens[0].Word != "龘龘" || tokens[0].IsDict {
		t.Fatalf("兜底应 bigram: %+v", tokens[0])
	}
	if tokens[1].Word != "龘" {
		t.Fatalf("奇数尾应单字: %+v", tokens[1])
	}
}

// TestSegmentMixed 拉丁/数字连串与标点切分
func TestSegmentMixed(t *testing.T) {
	tokens := Segment("侬好2026，shanghai！")
	var words []string
	for _, tk := range tokens {
		words = append(words, tk.Word)
	}
	joined := strings.Join(words, "|")
	if !strings.Contains(joined, "2026") || !strings.Contains(joined, "shanghai") {
		t.Fatalf("拉丁/数字连串应成词: %s", joined)
	}
	if strings.Contains(joined, "，") || strings.Contains(joined, "！") {
		t.Fatalf("标点不应成词: %s", joined)
	}
}

// TestWordFreq 词频统计降序
func TestWordFreq(t *testing.T) {
	freq := WordFreq("侬好侬好阿拉，阿拉白相")
	if len(freq) == 0 || freq[0].Word != "侬好" || freq[0].Count != 2 {
		t.Fatalf("「侬好」应 2 次居首: %+v", freq)
	}
	for i := 1; i < len(freq); i++ {
		if freq[i].Count > freq[i-1].Count {
			t.Fatal("词频应降序")
		}
	}
}
