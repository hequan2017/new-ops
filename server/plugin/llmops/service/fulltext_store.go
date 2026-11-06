// Package service llmops：全文检索库（要求(2) 四库之一）
// 口径：无 jieba 依赖——ASCII 按词元、CJK 按 bigram 切分；检索时内存统计词频，
// 简化 BM25(k1=1.5,b=0.75) 打分，IDF 在候选块集合内计算。
package service

import (
	"fmt"
	"strings"
	"unicode"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const ErrCodeFtTypeBad = 2244

// bm25 常量
const (
	bm25K1 = 1.5
	bm25B  = 0.75
)

// FtHit 全文检索命中
type FtHit struct {
	ChunkId uint    `json:"chunkId"`
	DocId   uint    `json:"docId"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
}

// Bigrams 分词：ASCII 连续段为词元(小写)；含 CJK 段切 bigram（单字保留）
func Bigrams(text string) []string {
	var tokens []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			tokens = append(tokens, cur.String())
			cur.Reset()
		}
	}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()

	var terms []string
	for _, tk := range tokens {
		runes := []rune(strings.ToLower(tk))
		hasCJK := false
		for _, r := range runes {
			if unicode.Is(unicode.Han, r) {
				hasCJK = true
				break
			}
		}
		if !hasCJK || len(runes) == 1 {
			terms = append(terms, string(runes))
			continue
		}
		for i := 0; i+1 < len(runes); i++ {
			terms = append(terms, string(runes[i:i+2]))
		}
	}
	return terms
}

// TermFreqs 词频表
func TermFreqs(terms []string) map[string]int {
	freq := make(map[string]int, len(terms))
	for _, t := range terms {
		freq[t]++
	}
	return freq
}

// BM25 单文档 BM25 得分（纯函数，测试锚点）：queryTerms 为查询词元（已去重）
func BM25Score(queryTerms []string, docFreq map[string]int, docLen int, avgLen float64, df map[string]int, totalDocs int) float64 {
	if docLen <= 0 || len(docFreq) == 0 {
		return 0
	}
	avg := avgLen
	if avg <= 0 {
		avg = 1
	}
	var score float64
	seen := make(map[string]bool, len(queryTerms))
	for _, q := range queryTerms {
		if seen[q] {
			continue
		}
		seen[q] = true
		tf := docFreq[q]
		if tf == 0 {
			continue
		}
		n := df[q]
		if n <= 0 {
			n = 1
		}
		idf := 1.0 + float64(totalDocs-n+1)/float64(n) // 非负 IDF 变体
		tfPart := float64(tf) * (bm25K1 + 1) / (float64(tf) + bm25K1*(1-bm25B+bm25B*float64(docLen)/avg))
		score += idf * tfPart
	}
	return score
}

// IndexDocToFulltext 文档索引：切块→落库（重建先清旧块）→回填 chunkCount
func (s *LlmopsServiceStruct) IndexDocToFulltext(kbId, docId uint) (*IndexResult, error) {
	var kb model.LlmopsKnowledge
	if err := global.GVA_DB.First(&kb, kbId).Error; err != nil {
		return nil, fmt.Errorf("[%d]知识库不存在", ErrCodeKbNotFound)
	}
	if kb.Type != "fulltext" {
		return nil, fmt.Errorf("[%d]仅 fulltext 类型知识库支持全文索引（当前 %s）", ErrCodeFtTypeBad, kb.Type)
	}
	var doc model.LlmopsKnowledgeDoc
	if err := global.GVA_DB.First(&doc, docId).Error; err != nil {
		return nil, fmt.Errorf("[%d]文档不存在", ErrCodeKbDocMiss)
	}
	if doc.KbId != kbId {
		return nil, fmt.Errorf("[%d]文档不属于该知识库", ErrCodeKbDocMiss)
	}
	chunks := ChunkText(doc.Content, ChunkOptions{})
	if len(chunks) == 0 {
		return nil, fmt.Errorf("[%d]文档无可切块内容", ErrCodeVecNoChunks)
	}
	rows := make([]model.LlmopsFulltextChunk, len(chunks))
	for i := range chunks {
		rows[i] = model.LlmopsFulltextChunk{KbId: kbId, DocId: docId, ChunkIndex: chunks[i].Index, Content: chunks[i].Text}
	}
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("doc_id = ?", docId).Delete(&model.LlmopsFulltextChunk{}).Error; err != nil {
			return err
		}
		if err := tx.CreateInBatches(rows, 200).Error; err != nil {
			return err
		}
		return tx.Model(&model.LlmopsKnowledgeDoc{}).Where("id = ?", docId).
			Updates(map[string]any{"chunk_count": len(rows), "status": "indexed"}).Error
	})
	if err != nil {
		return nil, err
	}
	return &IndexResult{DocId: docId, ChunkCount: len(rows), Source: "bigram-bm25"}, nil
}

// SearchFulltext 全文检索：候选块内存词频 → BM25 → top-k
func (s *LlmopsServiceStruct) SearchFulltext(kbId uint, query string, topK int) ([]FtHit, error) {
	var kb model.LlmopsKnowledge
	if err := global.GVA_DB.First(&kb, kbId).Error; err != nil {
		return nil, fmt.Errorf("[%d]知识库不存在", ErrCodeKbNotFound)
	}
	if kb.Type != "fulltext" {
		return nil, fmt.Errorf("[%d]仅 fulltext 类型知识库支持全文检索（当前 %s）", ErrCodeFtTypeBad, kb.Type)
	}
	if topK <= 0 || topK > 50 {
		topK = 5
	}
	var chunks []model.LlmopsFulltextChunk
	if err := global.GVA_DB.Where("kb_id = ?", kbId).Find(&chunks).Error; err != nil {
		return nil, err
	}
	if len(chunks) == 0 {
		return []FtHit{}, nil
	}
	queryTerms := Bigrams(query)
	if len(queryTerms) == 0 {
		return []FtHit{}, nil
	}
	// 候选集统计
	freqs := make([]map[string]int, len(chunks))
	df := make(map[string]int, 128)
	var totalLen int
	for i := range chunks {
		terms := Bigrams(chunks[i].Content)
		freqs[i] = TermFreqs(terms)
		totalLen += len(terms)
		for t := range freqs[i] {
			df[t]++
		}
	}
	avgLen := float64(totalLen) / float64(len(chunks))
	hits := make([]FtHit, 0, topK)
	for i := range chunks {
		score := BM25Score(queryTerms, freqs[i], len(freqs[i]), avgLen, df, len(chunks))
		if score <= 0 {
			continue
		}
		hit := FtHit{ChunkId: chunks[i].ID, DocId: chunks[i].DocId, Content: chunks[i].Content, Score: score}
		if len(hits) < topK {
			hits = append(hits, hit)
			if len(hits) == topK {
				sortFtHits(hits)
			}
			continue
		}
		if score > hits[topK-1].Score {
			hits[topK-1] = hit
			sortFtHits(hits)
		}
	}
	return hits, nil
}

// sortFtHits 按 Score 降序（插入排序，topK≤50）
func sortFtHits(hits []FtHit) {
	for i := 1; i < len(hits); i++ {
		for j := i; j > 0 && hits[j].Score > hits[j-1].Score; j-- {
			hits[j], hits[j-1] = hits[j-1], hits[j]
		}
	}
}
