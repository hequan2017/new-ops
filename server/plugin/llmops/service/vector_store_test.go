// Package service llmops 向量库单测（余弦相似度/哈希向量确定性/归一化/来源标识）
package service

import (
	"math"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

func TestCosineSimilarity(t *testing.T) {
	if got := CosineSimilarity([]float64{1, 0}, []float64{1, 0}); math.Abs(got-1) > 1e-9 {
		t.Fatalf("平行应得 1: %v", got)
	}
	if got := CosineSimilarity([]float64{1, 0}, []float64{0, 1}); math.Abs(got) > 1e-9 {
		t.Fatalf("正交应得 0: %v", got)
	}
	if got := CosineSimilarity([]float64{1, 0}, []float64{-1, 0}); math.Abs(got+1) > 1e-9 {
		t.Fatalf("反向应得 -1: %v", got)
	}
	if got := CosineSimilarity([]float64{0, 0}, []float64{1, 1}); got != 0 {
		t.Fatalf("零向量应得 0: %v", got)
	}
	if got := CosineSimilarity([]float64{1, 2}, []float64{1}); got != 0 {
		t.Fatalf("维度不齐应得 0: %v", got)
	}
}

// TestHashEmbed_Deterministic 同文本恒定、异文本不同、归一化、相近文本更相似
func TestHashEmbed_Deterministic(t *testing.T) {
	a1 := hashEmbed("白泽统一运维平台", hashEmbedDim)
	a2 := hashEmbed("白泽统一运维平台", hashEmbedDim)
	if len(a1) != hashEmbedDim {
		t.Fatalf("维度应为 %d: %d", hashEmbedDim, len(a1))
	}
	for i := range a1 {
		if a1[i] != a2[i] {
			t.Fatalf("同文本向量应恒定: idx %d", i)
		}
	}
	b := hashEmbed("另一个完全不同的文本内容", hashEmbedDim)
	same := true
	for i := range a1 {
		if a1[i] != b[i] {
			same = false
			break
		}
	}
	if same {
		t.Fatal("异文本向量不应全同")
	}
	var norm float64
	for _, v := range a1 {
		norm += v * v
	}
	if math.Abs(math.Sqrt(norm)-1) > 1e-9 {
		t.Fatalf("应 L2 归一: %v", math.Sqrt(norm))
	}
	q := hashEmbed("白泽统一运维平台下载", hashEmbedDim)
	simNear := CosineSimilarity(a1, q)
	irrel := hashEmbed("今天晚饭吃什么好呢", hashEmbedDim)
	simFar := CosineSimilarity(a1, irrel)
	if simNear <= simFar {
		t.Fatalf("相近文本应更相似: near=%v far=%v", simNear, simFar)
	}
}

// TestHashEmbed_Empty 空文本不 panic 且零向量；自定义维度
func TestHashEmbed_Empty(t *testing.T) {
	vec := hashEmbed("", hashEmbedDim)
	for _, v := range vec {
		if v != 0 {
			t.Fatalf("空文本应零向量: %v", v)
		}
	}
	if got := hashEmbed("x", 32); len(got) != 32 {
		t.Fatalf("自定义维度: %d", len(got))
	}
}

// TestVecSource 来源标识契约（检索同 Source 过滤依赖此格式）
func TestVecSource(t *testing.T) {
	withModel := &model.LlmopsKnowledge{EmbedModelId: 7}
	if got := vecSource(withModel); got != "embed:7" {
		t.Fatalf("embed 来源标识异常: %q", got)
	}
	noModel := &model.LlmopsKnowledge{EmbedModelId: 0}
	if got := vecSource(noModel); got != "hash-fallback" {
		t.Fatalf("哈希来源标识异常: %q", got)
	}
}
