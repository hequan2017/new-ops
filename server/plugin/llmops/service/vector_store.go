// Package service llmops：向量语义库（要求(2) 四库之一）
// 口径：向量化优先走库绑定的 embed 模型（B02 网关）；未绑模型或调用失败 → 确定性哈希向量
// fallback（fnv 双哈希映射 dim 桶 + L2 归一，同一文本恒定），Source 落库标注；
// 检索只对同 Source 分块计算余弦（跨向量空间不可比）。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeVecTypeBad  = 2242
	ErrCodeVecNoChunks = 2243
	hashEmbedDim       = 256
)

// VectorHit 向量检索命中
type VectorHit struct {
	ChunkId  uint    `json:"chunkId"`
	DocId    uint    `json:"docId"`
	Content  string  `json:"content"`
	Score    float64 `json:"score"`
}

// CosineSimilarity 余弦相似度（零向量返回 0）
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// hashEmbed 确定性哈希向量（纯函数，测试锚点）：fnv64a 字符级双哈希投桶 + 符号交替 + L2 归一
func hashEmbed(text string, dim int) []float64 {
	if dim <= 0 {
		dim = hashEmbedDim
	}
	vec := make([]float64, dim)
	for i, r := range text {
		h := fnv.New64a()
		_, _ = h.Write([]byte(string(r)))
		v := h.Sum64()
		idx := int(v % uint64(dim))
		sign := 1.0
		if (v>>32)&1 == 1 {
			sign = -1.0
		}
		// 位置衰减权重：越早出现的字符权重略高（保序信息）
		vec[idx] += sign * (1.0 + 1.0/(1+float64(i%16)))
	}
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if norm == 0 {
		return vec
	}
	for i := range vec {
		vec[i] /= norm
	}
	return vec
}

// vecSource 期望向量来源标识
func vecSource(kb *model.LlmopsKnowledge) string {
	if kb.EmbedModelId > 0 {
		return fmt.Sprintf("embed:%d", kb.EmbedModelId)
	}
	return "hash-fallback"
}

// embedTexts 按库配置向量化（embed 模型失败时回退哈希并切换 Source）
func (s *LlmopsServiceStruct) embedTexts(kb *model.LlmopsKnowledge, texts []string) ([][]float64, string, error) {
	if kb.EmbedModelId > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), embedHTTPClient.Timeout)
		defer cancel()
		vecs, err := s.Embeddings(ctx, kb.EmbedModelId, texts)
		if err == nil {
			return vecs, vecSource(kb), nil
		}
		global.GVA_LOG.Error(fmt.Sprintf("llmops embed 模型调用失败回退哈希: %v", err))
	}
	out := make([][]float64, len(texts))
	for i, t := range texts {
		out[i] = hashEmbed(t, hashEmbedDim)
	}
	return out, "hash-fallback", nil
}

// IndexDocToVector 文档索引：切块→向量化→分块落库→回填 chunkCount
func (s *LlmopsServiceStruct) IndexDocToVector(kbId, docId uint) (*IndexResult, error) {
	var kb model.LlmopsKnowledge
	if err := global.GVA_DB.First(&kb, kbId).Error; err != nil {
		return nil, fmt.Errorf("[%d]知识库不存在", ErrCodeKbNotFound)
	}
	if kb.Type != "vector" {
		return nil, fmt.Errorf("[%d]仅 vector 类型知识库支持向量索引（当前 %s）", ErrCodeVecTypeBad, kb.Type)
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
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.Text
	}
	vecs, source, err := s.embedTexts(&kb, texts)
	if err != nil {
		return nil, err
	}
	rows := make([]model.LlmopsVectorChunk, len(chunks))
	for i := range chunks {
		vecJSON, _ := json.Marshal(vecs[i])
		rows[i] = model.LlmopsVectorChunk{
			KbId: kbId, DocId: docId, ChunkIndex: chunks[i].Index,
			Content: chunks[i].Text, VectorJSON: string(vecJSON),
			Dim: len(vecs[i]), Source: source,
		}
	}
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("doc_id = ?", docId).Delete(&model.LlmopsVectorChunk{}).Error; err != nil {
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
	return &IndexResult{DocId: docId, ChunkCount: len(rows), Dim: len(vecs[0]), Source: source}, nil
}

// IndexResult 索引结果
type IndexResult struct {
	DocId      uint   `json:"docId"`
	ChunkCount int    `json:"chunkCount"`
	Dim        int    `json:"dim"`
	Source     string `json:"source"`
}

// SearchVector 向量检索：query 向量化 → 同 Source 分块内存余弦 → top-k
func (s *LlmopsServiceStruct) SearchVector(kbId uint, query string, topK int) ([]VectorHit, string, error) {
	var kb model.LlmopsKnowledge
	if err := global.GVA_DB.First(&kb, kbId).Error; err != nil {
		return nil, "", fmt.Errorf("[%d]知识库不存在", ErrCodeKbNotFound)
	}
	if kb.Type != "vector" {
		return nil, "", fmt.Errorf("[%d]仅 vector 类型知识库支持向量检索（当前 %s）", ErrCodeVecTypeBad, kb.Type)
	}
	if topK <= 0 || topK > 50 {
		topK = 5
	}
	qVec, source, err := s.embedTexts(&kb, []string{query})
	if err != nil {
		return nil, "", err
	}
	var chunks []model.LlmopsVectorChunk
	if err := global.GVA_DB.Where("kb_id = ? AND source = ?", kbId, source).Find(&chunks).Error; err != nil {
		return nil, "", err
	}
	hits := make([]VectorHit, 0, topK)
	for i := range chunks {
		var vec []float64
		if json.Unmarshal([]byte(chunks[i].VectorJSON), &vec) != nil {
			continue
		}
		score := CosineSimilarity(qVec[0], vec)
		if len(hits) < topK {
			hits = append(hits, VectorHit{ChunkId: chunks[i].ID, DocId: chunks[i].DocId, Content: chunks[i].Content, Score: score})
			if len(hits) == topK {
				sortVectorHits(hits)
			}
			continue
		}
		if score > hits[topK-1].Score { // 顶堆替换（hits 已排序，末位最小）
			hits[topK-1] = VectorHit{ChunkId: chunks[i].ID, DocId: chunks[i].DocId, Content: chunks[i].Content, Score: score}
			sortVectorHits(hits)
		}
	}
	return hits, source, nil
}

// sortVectorHits 按 Score 降序（插入排序，topK≤50 足够）
func sortVectorHits(hits []VectorHit) {
	for i := 1; i < len(hits); i++ {
		for j := i; j > 0 && hits[j].Score > hits[j-1].Score; j-- {
			hits[j], hits[j-1] = hits[j-1], hits[j]
		}
	}
}
