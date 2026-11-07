// Package service llmops：知识库文档重复检测（备用池 P02——simhash 简化版）
// 口径：64 位 simhash——bigram 特征 → FNV64a 哈希按位加权累加（位为 1 加权、0 减权）→
// 符号定比特；海明距离 ≤3 判疑似重复（64 位下标准阈值）；纯函数全链可测。
package service

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	simhashHammingThreshold = 3
)

// Simhash64 文本指纹（纯函数，测试锚点）
func Simhash64(text string) uint64 {
	var bits [64]int
	for _, term := range Bigrams(text) {
		h := fnv.New64a()
		_, _ = h.Write([]byte(term))
		v := h.Sum64()
		for i := 0; i < 64; i++ {
			if v&(1<<uint(i)) != 0 {
				bits[i]++
			} else {
				bits[i]--
			}
		}
	}
	var out uint64
	for i := 0; i < 64; i++ {
		if bits[i] > 0 {
			out |= 1 << uint(i)
		}
	}
	return out
}

// HammingDistance 海明距离（纯函数）
func HammingDistance(a, b uint64) int {
	x := a ^ b
	n := 0
	for x != 0 {
		x &= x - 1
		n++
	}
	return n
}

// DupHit 重复检测结果
type DupHit struct {
	DocId   uint    `json:"docId"`
	Title   string  `json:"title"`
	Dist    int     `json:"dist"`
}

// DetectDocDuplicate 检测指定文档与库内既有文档的重复（同库，排除自身）
func (s *LlmopsServiceStruct) DetectDocDuplicate(kbId, docId uint) ([]DupHit, error) {
	var target model.LlmopsKnowledgeDoc
	if err := global.GVA_DB.First(&target, docId).Error; err != nil {
		return nil, fmt.Errorf("[%d]文档不存在", ErrCodeKbDocMiss)
	}
	if target.KbId != kbId {
		return nil, fmt.Errorf("[%d]文档不属于该知识库", ErrCodeKbDocMiss)
	}
	targetHash := Simhash64(target.Content)
	var docs []model.LlmopsKnowledgeDoc
	if err := global.GVA_DB.Where("kb_id = ? AND id <> ?", kbId, docId).Find(&docs).Error; err != nil {
		return nil, err
	}
	hits := []DupHit{}
	for _, d := range docs {
		if strings.TrimSpace(d.Content) == "" {
			continue
		}
		dist := HammingDistance(targetHash, Simhash64(d.Content))
		if dist <= simhashHammingThreshold {
			hits = append(hits, DupHit{DocId: d.ID, Title: d.Title, Dist: dist})
		}
	}
	return hits, nil
}
