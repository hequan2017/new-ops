// Package service llmops：混合检索编排（要求(2) 多路召回 + RRF 融合 + Rerank 精排）
// 口径：五路（向量/全文/问答/图谱/摘要）goroutine 并发召回，单路失败不拖垮整体（报告留痕）；
// RRF(k=60) 按内容去重融合；Rerank 模型可选（未配置跳过，保持 RRF 序）。
package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeRouteBad  = 2251
	rrfK             = 60
	rrfMaxPathSample = 3 // 图谱路取前 3 个起点节点
)

// HybridRoutes 全部召回路由白名单
var HybridRoutes = []string{"vector", "fulltext", "qa", "graph", "summary"}

// ValidRoute 路由白名单
func ValidRoute(r string) bool {
	for _, v := range HybridRoutes {
		if v == r {
			return true
		}
	}
	return false
}

// RetrievalHit 统一召回命中
type RetrievalHit struct {
	Route   string  `json:"route"`
	DocId   uint    `json:"docId"`
	ChunkId uint    `json:"chunkId"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
	Rrf     float64 `json:"rrf"`
}

// rrfFuse RRF 融合（纯函数，测试锚点）：按内容去重合并，score=Σ 1/(k+rank)，降序稳定
func rrfFuse(lists [][]RetrievalHit, k int) []RetrievalHit {
	if k <= 0 {
		k = rrfK
	}
	score := make(map[string]float64)
	best := make(map[string]RetrievalHit)
	order := make([]string, 0, 32)
	for _, list := range lists {
		for rank, h := range list {
			key := strings.TrimSpace(h.Content)
			if key == "" {
				key = fmt.Sprintf("%s:%d:%d", h.Route, h.DocId, h.ChunkId)
			}
			if _, seen := score[key]; !seen {
				order = append(order, key)
				cp := h
				best[key] = cp
			}
			score[key] += 1.0 / float64(k+rank+1)
		}
	}
	out := make([]RetrievalHit, 0, len(order))
	for _, key := range order {
		h := best[key]
		h.Rrf = score[key]
		out = append(out, h)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Rrf > out[j].Rrf })
	return out
}

// formatGraphPath 路径文本化（纯函数）：A -关系-> B -关系-> C
func formatGraphPath(path []PathStep) string {
	if len(path) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(path[0].FromName)
	for _, st := range path {
		sb.WriteString(" -")
		sb.WriteString(st.Relation)
		sb.WriteString("-> ")
		sb.WriteString(st.ToName)
	}
	return sb.String()
}

// SearchGraphRoute 图谱路召回：起点名模糊定位(≤3) → 1 跳路径文本化
func (s *LlmopsServiceStruct) SearchGraphRoute(kbId uint, query string) ([]RetrievalHit, error) {
	var nodes []model.LlmopsGraphNode
	if err := global.GVA_DB.Where("kb_id = ? AND name LIKE ?", kbId, "%"+EscapeLike(query)+"%").
		Limit(rrfMaxPathSample).Find(&nodes).Error; err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, nil
	}
	var edges []model.LlmopsGraphEdge
	if err := global.GVA_DB.Where("kb_id = ?", kbId).Find(&edges).Error; err != nil {
		return nil, err
	}
	adj := make(map[uint][]GraphEdgeRef, len(edges))
	for _, e := range edges {
		adj[e.FromNodeId] = append(adj[e.FromNodeId], GraphEdgeRef{To: e.ToNodeId, Relation: e.Relation})
	}
	names := map[uint]string{}
	var all []model.LlmopsGraphNode
	if err := global.GVA_DB.Where("kb_id = ?", kbId).Find(&all).Error; err == nil {
		for _, n := range all {
			names[n.ID] = n.Name
		}
	}
	hits := make([]RetrievalHit, 0, 8)
	for i, n := range nodes {
		for _, path := range BfsPaths(adj, n.ID, 1) {
			for j := range path {
				path[j].FromName = names[path[j].FromId]
				path[j].ToName = names[path[j].ToId]
			}
			hits = append(hits, RetrievalHit{
				Route: "graph", DocId: uint(n.ID), ChunkId: n.ID,
				Content: formatGraphPath(path), Score: 1.0 / float64(i+2),
			})
		}
	}
	return hits, nil
}

// qaRoute 问答路适配
func (s *LlmopsServiceStruct) qaRoute(kbId uint, query string, topK int) ([]RetrievalHit, error) {
	qaHits, err := s.SearchQa(kbId, query, topK, 0)
	if err != nil {
		return nil, err
	}
	out := make([]RetrievalHit, 0, len(qaHits))
	for _, h := range qaHits {
		out = append(out, RetrievalHit{Route: "qa", DocId: h.PairId, ChunkId: h.PairId,
			Content: h.Question + "\n答：" + h.Answer, Score: h.Score})
	}
	return out, nil
}

// HybridSearch 混合检索：并发多路召回 → RRF 融合 → Rerank 精排（可选）
// 返回 hits 与各路报告（ok/错误摘要）。
func (s *LlmopsServiceStruct) HybridSearch(kbId uint, query string, topK int, routes []string, rerankModelId uint) ([]RetrievalHit, map[string]string, error) {
	if topK <= 0 || topK > 50 {
		topK = 5
	}
	if len(routes) == 0 {
		routes = HybridRoutes
	}
	for _, r := range routes {
		if !ValidRoute(r) {
			return nil, nil, fmt.Errorf("[%d]不支持的召回路由: %s", ErrCodeRouteBad, r)
		}
	}
	var mu sync.Mutex
	var collected [][]RetrievalHit
	report := map[string]string{}
	var wg sync.WaitGroup

	run := func(name string, fn func() ([]RetrievalHit, error)) {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				mu.Lock()
				report[name] = fmt.Sprintf("panic: %v", r)
				mu.Unlock()
			}
		}()
		hits, err := fn()
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			report[name] = err.Error()
			return
		}
		collected = append(collected, hits)
		report[name] = "ok"
	}

	for _, route := range routes {
		r := route
		switch r {
		case "vector":
			wg.Add(1)
			go run("vector", func() ([]RetrievalHit, error) {
				vh, _, err := s.SearchVector(kbId, query, topK)
				if err != nil {
					return nil, err
				}
				out := make([]RetrievalHit, 0, len(vh))
				for _, h := range vh {
					out = append(out, RetrievalHit{Route: "vector", DocId: h.DocId, ChunkId: h.ChunkId, Content: h.Content, Score: h.Score})
				}
				return out, nil
			})
		case "fulltext":
			wg.Add(1)
			go run("fulltext", func() ([]RetrievalHit, error) {
				fh, err := s.SearchFulltext(kbId, query, topK)
				if err != nil {
					return nil, err
				}
				out := make([]RetrievalHit, 0, len(fh))
				for _, h := range fh {
					out = append(out, RetrievalHit{Route: "fulltext", DocId: h.DocId, ChunkId: h.ChunkId, Content: h.Content, Score: h.Score})
				}
				return out, nil
			})
		case "qa":
			wg.Add(1)
			go run("qa", func() ([]RetrievalHit, error) { return s.qaRoute(kbId, query, topK) })
		case "graph":
			wg.Add(1)
			go run("graph", func() ([]RetrievalHit, error) { return s.SearchGraphRoute(kbId, query) })
		case "summary":
			wg.Add(1)
			go run("summary", func() ([]RetrievalHit, error) {
				sh, err := s.SearchSummaries(kbId, query, topK)
				if err != nil {
					return nil, err
				}
				out := make([]RetrievalHit, 0, len(sh))
				for _, h := range sh {
					out = append(out, RetrievalHit{Route: "summary", DocId: h.DocId, ChunkId: h.ChunkId, Content: h.Content, Score: h.Score})
				}
				return out, nil
			})
		}
	}
	wg.Wait()

	fused := rrfFuse(collected, rrfK)

	// Rerank 精排（可选）：取 RRF 前 20 候选重排
	if rerankModelId > 0 && len(fused) > 0 {
		cand := fused
		if len(cand) > 20 {
			cand = cand[:20]
		}
		docs := make([]string, len(cand))
		for i := range cand {
			docs[i] = cand[i].Content
		}
		ctx, cancel := context.WithTimeout(context.Background(), embedHTTPClient.Timeout)
		defer cancel()
		if scores, err := s.Rerank(ctx, rerankModelId, query, docs); err == nil {
			reordered := make([]RetrievalHit, 0, len(scores))
			used := make(map[int]bool, len(scores))
			for _, sc := range scores {
				if sc.Index >= 0 && sc.Index < len(cand) && !used[sc.Index] {
					used[sc.Index] = true
					h := cand[sc.Index]
					h.Score = sc.Score
					reordered = append(reordered, h)
				}
			}
			for i := range cand { // 未被重排覆盖的候选按原序补尾
				if !used[i] {
					reordered = append(reordered, cand[i])
				}
			}
			fused = reordered
		} else {
			report["rerank"] = err.Error()
		}
	}
	if len(fused) > topK {
		fused = fused[:topK]
	}
	return fused, report, nil
}
