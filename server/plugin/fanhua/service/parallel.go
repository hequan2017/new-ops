// Package service fanhua：双语平行语料（需求 3.4-4/15）
// 口径：对照检索=普→沪 编辑相似度 top-k（编辑距离纯标准库 rune 实现，对齐 llmops C05 口径自实现）；
// 种子 ≥20 条沪-普-英三元组（幂等）。
package service

import (
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/csvx"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodeParaNotFound = 2604
	ErrCodeParaField    = 2605
)

// EditDistance 编辑距离（rune 级双行 DP，纯函数）
func EditDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	curr := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		curr[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			curr[j] = min3(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(rb)]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

// EditSimilarity 编辑相似度 = 1 - dist/maxLen（纯函数，测试锚点）
func EditSimilarity(a, b string) float64 {
	la, lb := len([]rune(a)), len([]rune(b))
	if la == 0 && lb == 0 {
		return 1
	}
	maxLen := la
	if lb > maxLen {
		maxLen = lb
	}
	return 1 - float64(EditDistance(a, b))/float64(maxLen)
}

// paraHit 检索命中项
type paraHit struct {
	model.FanhuaParallel
	Score float64 `json:"score"`
}

// rankParallel 对照排序（纯函数，测试锚点）：普→沪相似度取 max(Pu 命中, Sh 命中)。
// 编辑相似度与 text.Bigrams/Score 的 bigram 集合计数不等价（N03 口径为连续子串编辑距离，
// 可度量长句渐进相似；bigram 集合计数无此单调性），保留本地实现与本地排序。
func rankParallel(rows []model.FanhuaParallel, query string, topK int) []paraHit {
	hits := make([]paraHit, 0, len(rows))
	for _, r := range rows {
		score := EditSimilarity(query, r.Pu)
		if s2 := EditSimilarity(query, r.Sh); s2 > score {
			score = s2
		}
		hits = append(hits, paraHit{FanhuaParallel: r, Score: float64(int(score*1e4)) / 1e4})
	}
	for i := 1; i < len(hits); i++ {
		for j := i; j > 0 && hits[j].Score > hits[j-1].Score; j-- {
			hits[j], hits[j-1] = hits[j-1], hits[j]
		}
	}
	if topK > 0 && len(hits) > topK {
		hits = hits[:topK]
	}
	return hits
}

// validateParallel 保存校验（纯函数）
func validateParallel(in *model.FanhuaParallel) error {
	if strings.TrimSpace(in.Sh) == "" || strings.TrimSpace(in.Pu) == "" {
		return errx.New(ErrCodeParaField, "沪语与普通话句子必填")
	}
	return nil
}

// SearchParallel 对照检索（普→沪 top-k）
func (s *FanhuaService) SearchParallel(query string, topK int) ([]paraHit, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errx.New(ErrCodeParaField, "检索语句必填")
	}
	if topK < 1 || topK > 50 {
		topK = 5
	}
	rows, err := s.repo.ParallelsScan(s.dbh())
	if err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "扫描平行语料失败", err)
	}
	return rankParallel(rows, query, topK), nil
}

// SaveParallel 新建/更新平行语料
func (s *FanhuaService) SaveParallel(in *model.FanhuaParallel) (*model.FanhuaParallel, error) {
	if err := validateParallel(in); err != nil {
		return nil, err
	}
	db := s.dbh()
	if in.ID > 0 {
		exist, err := s.repo.FindParallel(db, in.ID)
		if err != nil {
			return nil, errx.New(ErrCodeParaNotFound, "平行语料不存在")
		}
		if err := s.repo.UpdateParallelColumns(db, exist.ID, map[string]any{
			"sh": in.Sh, "pu": in.Pu, "en": in.En, "source": in.Source,
		}); err != nil {
			return nil, errx.Wrap(ErrCodeFanhuaStore, "更新平行语料失败", err)
		}
		return exist, nil
	}
	if err := s.repo.CreateParallel(db, in); err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "创建平行语料失败", err)
	}
	return in, nil
}

// DeleteParallel 删除平行语料（软删除）
func (s *FanhuaService) DeleteParallel(id uint) error {
	rows, err := s.repo.DeleteParallel(s.dbh(), id)
	if err != nil {
		return errx.Wrap(ErrCodeFanhuaStore, "删除平行语料失败", err)
	}
	if rows == 0 {
		return errx.New(ErrCodeParaNotFound, "平行语料不存在")
	}
	return nil
}

// GetParallels 平行语料列表（关键字过滤，倒序分页）
func (s *FanhuaService) GetParallels(keyword string, page, pageSize int) ([]model.FanhuaParallel, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	list, total, err := s.repo.ParallelList(s.dbh(), strings.TrimSpace(keyword), page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeFanhuaStore, "查询平行语料失败", err)
	}
	return list, total, nil
}

// parallelsToCSV 平行语料 CSV（纯函数，测试锚点，I07 口径）：复用 csvx.Builder
// 全字段清洗（换行折空格先于逗号转全角），与原实现逐字段 clean 语义等价。
func parallelsToCSV(list []model.FanhuaParallel) string {
	b := csvx.New("沪语", "普通话", "英语", "来源")
	for _, r := range list {
		b.Row(r.Sh, r.Pu, r.En, r.Source)
	}
	return b.String()
}

// ExportParallels 平行语料 CSV 导出（来源过滤，倒序 ≤5000 行）
func (s *FanhuaService) ExportParallels(source string) (string, error) {
	list, err := s.repo.ParallelsForExport(s.dbh(), source)
	if err != nil {
		return "", errx.Wrap(ErrCodeFanhuaStore, "扫描平行语料导出失败", err)
	}
	return parallelsToCSV(list), nil
}
