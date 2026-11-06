// Package service llmops：统一检索入口（要求(2) 收口）——多库多模式检索 + 调用留痕
// 口径：mode=hybrid 走 C08 五路编排；单路由模式按 mode 定路；多库结果再经 RRF 一次融合；
// 成败均落 LlmopsSearchLog（时延/命中数/各路报告）。
package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// UnifiedSearchIn 统一检索请求
type UnifiedSearchIn struct {
	KbIds         []uint `json:"kbIds" binding:"required,min=1"`
	Query         string `json:"query" binding:"required"`
	Mode          string `json:"mode"` // hybrid(默认)/vector/fulltext/qa/graph/summary
	TopK          int    `json:"topK"`
	RerankModelId uint   `json:"rerankModelId"`
}

// ValidSearchMode 检索模式白名单（hybrid + 五单路由）
func ValidSearchMode(m string) bool {
	if m == "" || m == "hybrid" {
		return true
	}
	return ValidRoute(m)
}

// routeHit 单路由对单库检索（统一适配）
func (s *LlmopsServiceStruct) routeHit(mode string, kbId uint, query string, topK int) ([]RetrievalHit, error) {
	switch mode {
	case "vector":
		vh, _, err := s.SearchVector(kbId, query, topK)
		if err != nil {
			return nil, err
		}
		out := make([]RetrievalHit, 0, len(vh))
		for _, h := range vh {
			out = append(out, RetrievalHit{Route: "vector", DocId: h.DocId, ChunkId: h.ChunkId, Content: h.Content, Score: h.Score})
		}
		return out, nil
	case "fulltext":
		fh, err := s.SearchFulltext(kbId, query, topK)
		if err != nil {
			return nil, err
		}
		out := make([]RetrievalHit, 0, len(fh))
		for _, h := range fh {
			out = append(out, RetrievalHit{Route: "fulltext", DocId: h.DocId, ChunkId: h.ChunkId, Content: h.Content, Score: h.Score})
		}
		return out, nil
	case "qa":
		return s.qaRoute(kbId, query, topK)
	case "graph":
		return s.SearchGraphRoute(kbId, query)
	default: // summary
		sh, err := s.SearchSummaries(kbId, query, topK)
		if err != nil {
			return nil, err
		}
		out := make([]RetrievalHit, 0, len(sh))
		for _, h := range sh {
			out = append(out, RetrievalHit{Route: "summary", DocId: h.DocId, ChunkId: h.ChunkId, Content: h.Content, Score: h.Score})
		}
		return out, nil
	}
}

// UnifiedSearch 统一检索（多库 → 融合 → topK，成败均留痕）
func (s *LlmopsServiceStruct) UnifiedSearch(in *UnifiedSearchIn) ([]RetrievalHit, map[string]string, error) {
	if in.Mode == "" {
		in.Mode = "hybrid"
	}
	if !ValidSearchMode(in.Mode) {
		return nil, nil, fmt.Errorf("[%d]不支持的检索模式: %s", ErrCodeRouteBad, in.Mode)
	}
	if in.TopK <= 0 || in.TopK > 50 {
		in.TopK = 5
	}
	started := time.Now()
	var lists [][]RetrievalHit
	report := map[string]string{}
	var searchErr error

	if in.Mode == "hybrid" {
		for _, kbId := range in.KbIds {
			hits, rep, err := s.HybridSearch(kbId, in.Query, in.TopK, nil, in.RerankModelId)
			for k, v := range rep {
				report[fmt.Sprintf("kb%d/%s", kbId, k)] = v
			}
			if err != nil {
				searchErr = err
				break
			}
			lists = append(lists, hits)
		}
	} else {
		for _, kbId := range in.KbIds {
			hits, err := s.routeHit(in.Mode, kbId, in.Query, in.TopK)
			if err != nil {
				searchErr = err
				report[fmt.Sprintf("kb%d", kbId)] = err.Error()
				break
			}
			lists = append(lists, hits)
		}
	}

	latency := time.Since(started).Milliseconds()
	var out []RetrievalHit
	if searchErr == nil {
		out = rrfFuse(lists, rrfK)
		if len(out) > in.TopK {
			out = out[:in.TopK]
		}
	}
	s.recordSearchLog(in, len(out), latency, report, searchErr)
	if searchErr != nil {
		return nil, report, searchErr
	}
	return out, report, nil
}

// recordSearchLog 检索留痕（失败仅记日志不阻断）
func (s *LlmopsServiceStruct) recordSearchLog(in *UnifiedSearchIn, hitCount int, latencyMs int64, report map[string]string, searchErr error) {
	kbIds := make([]string, len(in.KbIds))
	for i, id := range in.KbIds {
		kbIds[i] = strconv.FormatUint(uint64(id), 10)
	}
	repStr := ""
	if len(report) > 0 {
		parts := make([]string, 0, len(report))
		for k, v := range report {
			parts = append(parts, k+"="+v)
		}
		repStr = truncateSample(strings.Join(parts, ";"))
		if len([]rune(repStr)) > 500 {
			repStr = string([]rune(repStr)[:500])
		}
	}
	log := model.LlmopsSearchLog{
		KbIds: strings.Join(kbIds, ","), Query: truncateSample(in.Query), Mode: in.Mode,
		TopK: in.TopK, HitCount: hitCount, LatencyMs: latencyMs,
		RerankModelId: in.RerankModelId, RouteReport: repStr,
	}
	if searchErr != nil {
		e := truncateSample(searchErr.Error())
		if len([]rune(e)) > 500 {
			e = string([]rune(e)[:500])
		}
		log.ErrMsg = e
	}
	if err := global.GVA_DB.Create(&log).Error; err != nil {
		global.GVA_LOG.Error(fmt.Sprintf("llmops 检索日志写入失败: %v", err))
	}
}

// GetSearchLogs 检索日志列表（倒序分页）
func (s *LlmopsServiceStruct) GetSearchLogs(page, pageSize int) ([]model.LlmopsSearchLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.LlmopsSearchLog{})
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.LlmopsSearchLog
	err := db.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
