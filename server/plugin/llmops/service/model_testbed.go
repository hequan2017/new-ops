// Package service llmops：模型在线测试台（chat/embeddings/rerank 三协议 + 调用日志留痕）
// rerank 协议约定（非 OpenAI 标准）：POST {base}/rerank，请求 {model,query,documents}，
// 响应 {"results":[{"index":0,"score":0.98},...]}（对齐 Jina/Cohere 风格，注明口径）。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const ErrCodeRerank = 2234

// RerankScore 重排得分（Document 回填原文便于前端展示）
type RerankScore struct {
	Index    int     `json:"index"`
	Score    float64 `json:"score"`
	Document string  `json:"document,omitempty"`
}

// TestProtocolIn 在线测试请求
type TestProtocolIn struct {
	ModelId   uint           `json:"modelId" binding:"required"`
	Protocol  string         `json:"protocol" binding:"required"` // chat/embed/rerank
	Messages  []ChatMessage  `json:"messages"`
	Input     []string       `json:"input"`
	Query     string         `json:"query"`
	Documents []string       `json:"documents"`
	Opts      map[string]any `json:"opts"`
}

// parseRerankResponse 解析 rerank 响应（纯函数，测试锚点）
func parseRerankResponse(data []byte) ([]RerankScore, error) {
	var parsed struct {
		Results []struct {
			Index int     `json:"index"`
			Score float64 `json:"score"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("[%d]rerank 响应解析失败: %w", ErrCodeGatewayParse, err)
	}
	if len(parsed.Results) == 0 {
		return nil, fmt.Errorf("[%d]rerank 响应解析失败: results 为空", ErrCodeRerank)
	}
	out := make([]RerankScore, len(parsed.Results))
	for i, r := range parsed.Results {
		out[i] = RerankScore{Index: r.Index, Score: r.Score}
	}
	return out, nil
}

// Rerank 重排调用（rerank 能力入口校验）
func (s *LlmopsServiceStruct) Rerank(ctx context.Context, modelID uint, query string, documents []string) ([]RerankScore, error) {
	if query == "" || len(documents) == 0 {
		return nil, fmt.Errorf("[%d]query 与 documents 不能为空", ErrCodeRerank)
	}
	m, apiKey, err := s.loadModelFor(modelID, "rerank")
	if err != nil {
		return nil, err
	}
	data, err := gatewayPost(ctx, embedHTTPClient, gatewayEndpoint(m.BaseUrl, "/rerank"), apiKey, map[string]any{
		"model": m.Name, "query": query, "documents": documents,
	})
	if err != nil {
		return nil, err
	}
	scores, err := parseRerankResponse(data)
	if err != nil {
		return nil, err
	}
	for i := range scores { // 回填原文
		if scores[i].Index >= 0 && scores[i].Index < len(documents) {
			scores[i].Document = documents[scores[i].Index]
		}
	}
	return scores, nil
}

// recordCallLog 调用日志落库（失败仅记日志不阻断）
func (s *LlmopsServiceStruct) recordCallLog(m *model.LlmopsModel, protocol string, started time.Time, usage *ChatUsage, callErr error) {
	log := model.LlmopsModelCallLog{
		ModelId: m.ID, ModelName: m.Name, Protocol: protocol,
		LatencyMs: time.Since(started).Milliseconds(), Ok: callErr == nil,
	}
	if usage != nil {
		log.PromptTokens, log.CompletionTokens, log.TotalTokens = usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens
	}
	if callErr != nil {
		log.Error = truncateSample(callErr.Error())
		if len([]rune(log.Error)) > 500 {
			log.Error = string([]rune(log.Error)[:500])
		}
	}
	if err := global.GVA_DB.Create(&log).Error; err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("llmops 调用日志写入失败: %v", err))
	}
}

// TestModelProtocol 在线测试（三协议统一入口，成败均留痕）
func (s *LlmopsServiceStruct) TestModelProtocol(in *TestProtocolIn) (any, error) {
	if !ValidCapability(in.Protocol) {
		return nil, fmt.Errorf("[%d]不支持的协议: %s", ErrCodeModelCapBad, in.Protocol)
	}
	m, err := s.GetModel(in.ModelId)
	if err != nil {
		return nil, err
	}
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 130*time.Second)
	defer cancel()

	switch in.Protocol {
	case "chat":
		content, usage, err := s.ChatCompletion(ctx, in.ModelId, in.Messages, in.Opts)
		s.recordCallLog(m, "chat", started, usage, err)
		if err != nil {
			return nil, err
		}
		return map[string]any{"content": content, "usage": usage}, nil
	case "embed":
		vecs, err := s.Embeddings(ctx, in.ModelId, in.Input)
		s.recordCallLog(m, "embed", started, nil, err)
		if err != nil {
			return nil, err
		}
		return map[string]any{"dim": len(vecs[0]), "count": len(vecs)}, nil
	default: // rerank
		scores, err := s.Rerank(ctx, in.ModelId, in.Query, in.Documents)
		s.recordCallLog(m, "rerank", started, nil, err)
		if err != nil {
			return nil, err
		}
		return map[string]any{"results": scores}, nil
	}
}

// GetModelCallLogs 调用日志列表（模型/成败过滤，倒序分页）
func (s *LlmopsServiceStruct) GetModelCallLogs(page, pageSize int, modelId uint, okFilter string) ([]model.LlmopsModelCallLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.LlmopsModelCallLog{})
	if modelId > 0 {
		db = db.Where("model_id = ?", modelId)
	}
	if okFilter == "true" || okFilter == "false" {
		db = db.Where("ok = ?", okFilter == "true")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.LlmopsModelCallLog
	err := db.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
