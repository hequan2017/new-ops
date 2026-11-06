// Package service llmops：问答对库检索（要求(2) 四库之一）
// 口径：候选对 = 库内文档(AssetId>0)关联的问答对；归一键精确命中优先(1.0)，
// 其余按编辑相似度(1-dist/maxLen)降序，低于阈值过滤；相似度计算为纯函数锚点。
package service

import (
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeQaKbTypeBad = 2245
	qaDefaultMinScore  = 0.35
)

// QaHit 问答检索命中
type QaHit struct {
	PairId   uint    `json:"pairId"`
	Question string  `json:"question"`
	Answer   string  `json:"answer"`
	Score    float64 `json:"score"`
	Exact    bool    `json:"exact"`
}

// EditDistance 编辑距离（Levenshtein DP，纯函数）
func EditDistance(a, b []rune) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			m := del
			if ins < m {
				m = ins
			}
			if sub < m {
				m = sub
			}
			curr[j] = m
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

// QaSimilarity 归一编辑相似度（0-1；1=完全一致）
func QaSimilarity(query, question string) float64 {
	q := []rune(NormalizeQaKey(query))
	t := []rune(NormalizeQaKey(question))
	if len(q) == 0 && len(t) == 0 {
		return 1
	}
	maxLen := len(q)
	if len(t) > maxLen {
		maxLen = len(t)
	}
	if maxLen == 0 {
		return 0
	}
	return 1 - float64(EditDistance(q, t))/float64(maxLen)
}

// rankQaPairs 候选问答对排序（纯函数，测试锚点）：精确归一命中 1.0 优先，其余相似度降序过滤
func rankQaPairs(query string, pairs []model.LlmopsQaPair, minScore float64) []QaHit {
	hits := make([]QaHit, 0, len(pairs))
	qKey := NormalizeQaKey(query)
	for i := range pairs {
		p := &pairs[i]
		score := QaSimilarity(query, p.Question)
		exact := qKey != "" && qKey == NormalizeQaKey(p.Question)
		if exact {
			score = 1
		}
		if score < minScore {
			continue
		}
		hits = append(hits, QaHit{PairId: p.ID, Question: p.Question, Answer: p.Answer, Score: score, Exact: exact})
	}
	// 精确优先，其次分数降序，再次 pairId 升序（稳定）
	for i := 1; i < len(hits); i++ {
		for j := i; j > 0; j-- {
			a, b := hits[j], hits[j-1]
			if a.Exact != b.Exact {
				if !b.Exact && a.Exact {
					hits[j], hits[j-1] = hits[j-1], hits[j]
				}
				continue
			}
			if a.Score > b.Score || (a.Score == b.Score && a.PairId < b.PairId) {
				hits[j], hits[j-1] = hits[j-1], hits[j]
			} else {
				break
			}
		}
	}
	return hits
}

// SearchQa 问答检索：库内文档关联问答对 → rank → topK
func (s *LlmopsServiceStruct) SearchQa(kbId uint, query string, topK int, minScore float64) ([]QaHit, error) {
	var kb model.LlmopsKnowledge
	if err := global.GVA_DB.First(&kb, kbId).Error; err != nil {
		return nil, fmt.Errorf("[%d]知识库不存在", ErrCodeKbNotFound)
	}
	if kb.Type != "qa" {
		return nil, fmt.Errorf("[%d]仅 qa 类型知识库支持问答检索（当前 %s）", ErrCodeQaKbTypeBad, kb.Type)
	}
	if topK <= 0 || topK > 50 {
		topK = 5
	}
	if minScore <= 0 || minScore > 1 {
		minScore = qaDefaultMinScore
	}
	var pairs []model.LlmopsQaPair
	err := global.GVA_DB.Where("asset_id IN (?)",
		global.GVA_DB.Model(&model.LlmopsKnowledgeDoc{}).Select("asset_id").Where("kb_id = ? AND asset_id > 0", kbId),
	).Order("id").Find(&pairs).Error
	if err != nil {
		return nil, err
	}
	hits := rankQaPairs(query, pairs, minScore)
	if len(hits) > topK {
		hits = hits[:topK]
	}
	return hits, nil
}
