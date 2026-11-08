// Package service fanhua：平行语料单测（纯函数，无 DB）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// TestEditDistance 编辑距离锚点
func TestEditDistance(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"侬好", "侬好", 0},
		{"侬好", "你好", 1},
		{"kitten", "sitting", 3},
		{"", "abc", 3},
	}
	for _, c := range cases {
		if got := EditDistance(c.a, c.b); got != c.want {
			t.Fatalf("EditDistance(%q,%q)=%d 期望 %d", c.a, c.b, got, c.want)
		}
	}
}

// TestEditSimilarity 相似度：全等 1、无关趋 0、单调
func TestEditSimilarity(t *testing.T) {
	if s := EditSimilarity("侬好", "侬好"); s != 1 {
		t.Fatalf("全等应为 1: %v", s)
	}
	close1 := EditSimilarity("饭吃过了伐", "饭吃过了没")
	far := EditSimilarity("饭吃过了伐", "今天天气非常好")
	if close1 <= far {
		t.Fatalf("相近句相似度应更高: %v vs %v", close1, far)
	}
	if s := EditSimilarity("", ""); s != 1 {
		t.Fatalf("空串全等应为 1: %v", s)
	}
}

// TestRankParallel 对照排序：Pu 命中优先、topK 截取
func TestRankParallel(t *testing.T) {
	rows := []model.FanhuaParallel{
		{Sh: "明朝见", Pu: "明天见"},
		{Sh: "侬好", Pu: "你好"},
		{Sh: "今朝天气交关好", Pu: "今天天气非常好"},
	}
	hits := rankParallel(rows, "你好呀", 2)
	if len(hits) != 2 {
		t.Fatalf("topK 应截取 2: %d", len(hits))
	}
	if hits[0].Pu != "你好" {
		t.Fatalf("首位应为「你好」: %s", hits[0].Pu)
	}
	if hits[0].Score <= hits[1].Score {
		t.Fatalf("应降序: %v %v", hits[0].Score, hits[1].Score)
	}
}

// TestValidateParallel 保存校验
func TestValidateParallel(t *testing.T) {
	if err := validateParallel(&model.FanhuaParallel{Sh: "侬好", Pu: "你好"}); err != nil {
		t.Fatalf("合法应通过: %v", err)
	}
	if err := validateParallel(&model.FanhuaParallel{Sh: "", Pu: "你好"}); err == nil {
		t.Fatal("沪语空应报错")
	}
}
