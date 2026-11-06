// Package service llmops：摘要检索（要求(2) 多路召回之一）——文档摘要异步生成 + 摘要内检索
// 口径：生成走 chat 模型（goroutine 异步，panic recover 兜底）；长文取前 8000 rune 送摘要；
// 摘要检索复用 C04 bigram+BM25 打分；同文档重复生成覆盖旧任务。
package service

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码与口径常量
const (
	ErrCodeSummaryNotReady = 2250
	summaryInputMaxRune    = 8000
)

// truncateForSummary 摘要输入截断（长文取前段，纯函数）
func truncateForSummary(content string, maxRunes int) string {
	r := []rune(strings.TrimSpace(content))
	if len(r) <= maxRunes {
		return string(r)
	}
	return string(r[:maxRunes]) + "\n（内容过长，已截取前段生成摘要）"
}

// buildSummaryPrompt 摘要 prompt（纯函数，测试锚点）
func buildSummaryPrompt(content string) []ChatMessage {
	return []ChatMessage{
		{Role: "system", Content: "你是知识库摘要助手。用 3-5 句话概括文档核心内容，保留关键实体与结论，不要输出无关寒暄。"},
		{Role: "user", Content: content},
	}
}

// CreateSummaryTask 创建摘要生成任务（异步 goroutine 执行；同文档覆盖旧任务）
func (s *LlmopsServiceStruct) CreateSummaryTask(kbId, docId, modelId uint) (*model.LlmopsDocSummary, error) {
	var kb model.LlmopsKnowledge
	if err := global.GVA_DB.First(&kb, kbId).Error; err != nil {
		return nil, fmt.Errorf("[%d]知识库不存在", ErrCodeKbNotFound)
	}
	var doc model.LlmopsKnowledgeDoc
	if err := global.GVA_DB.Where("id = ? AND kb_id = ?", docId, kbId).First(&doc).Error; err != nil {
		return nil, fmt.Errorf("[%d]文档不存在或不属于该库", ErrCodeKbDocMiss)
	}
	task := &model.LlmopsDocSummary{KbId: kbId, DocId: docId, Status: "pending", ModelId: modelId}
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("doc_id = ?", docId).Delete(&model.LlmopsDocSummary{}).Error; err != nil {
			return err
		}
		return tx.Create(task).Error
	})
	if err != nil {
		return nil, err
	}
	go s.runSummaryTask(task.ID)
	return task, nil
}

// runSummaryTask 执行摘要生成（异步；panic 兜底置 failed）
func (s *LlmopsServiceStruct) runSummaryTask(taskId uint) {
	defer func() {
		if r := recover(); r != nil {
			global.GVA_LOG.Error(fmt.Sprintf("llmops 摘要任务 panic task=%d: %v", taskId, r))
			global.GVA_DB.Model(&model.LlmopsDocSummary{}).Where("id = ?", taskId).
				Updates(map[string]any{"status": "failed", "err_msg": "internal panic"})
		}
	}()
	var task model.LlmopsDocSummary
	if err := global.GVA_DB.First(&task, taskId).Error; err != nil {
		return
	}
	global.GVA_DB.Model(&task).Update("status", "running")

	var doc model.LlmopsKnowledgeDoc
	if err := global.GVA_DB.First(&doc, task.DocId).Error; err != nil {
		global.GVA_DB.Model(&task).Updates(map[string]any{"status": "failed", "err_msg": "文档不存在"})
		return
	}
	content, _, err := s.ChatCompletion(context.Background(), task.ModelId, buildSummaryPrompt(truncateForSummary(doc.Content, summaryInputMaxRune)), nil)
	if err != nil {
		global.GVA_DB.Model(&task).Updates(map[string]any{"status": "failed", "err_msg": truncateSample(err.Error())})
		return
	}
	global.GVA_DB.Model(&task).Updates(map[string]any{"status": "done", "summary": content, "err_msg": ""})
}

// GetDocSummary 查询文档摘要任务状态/结果
func (s *LlmopsServiceStruct) GetDocSummary(kbId, docId uint) (*model.LlmopsDocSummary, error) {
	var task model.LlmopsDocSummary
	err := global.GVA_DB.Where("kb_id = ? AND doc_id = ?", kbId, docId).First(&task).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("[%d]该文档尚无摘要任务", ErrCodeSummaryNotReady)
		}
		return nil, err
	}
	return &task, nil
}

// SearchSummaries 摘要检索（复用 bigram+BM25 打分，仅返回 done 摘要）
func (s *LlmopsServiceStruct) SearchSummaries(kbId uint, query string, topK int) ([]FtHit, error) {
	if topK <= 0 || topK > 50 {
		topK = 5
	}
	var tasks []model.LlmopsDocSummary
	if err := global.GVA_DB.Where("kb_id = ? AND status = 'done'", kbId).Find(&tasks).Error; err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return []FtHit{}, nil
	}
	queryTerms := Bigrams(query)
	if len(queryTerms) == 0 {
		return []FtHit{}, nil
	}
	freqs := make([]map[string]int, len(tasks))
	df := make(map[string]int, 64)
	var totalLen int
	for i := range tasks {
		terms := Bigrams(tasks[i].Summary)
		freqs[i] = TermFreqs(terms)
		totalLen += len(terms)
		for t := range freqs[i] {
			df[t]++
		}
	}
	avgLen := float64(totalLen) / float64(len(tasks))
	hits := make([]FtHit, 0, topK)
	for i := range tasks {
		score := BM25Score(queryTerms, freqs[i], len(freqs[i]), avgLen, df, len(tasks))
		if score <= 0 {
			continue
		}
		hit := FtHit{ChunkId: tasks[i].DocId, DocId: tasks[i].DocId, Content: tasks[i].Summary, Score: score}
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
