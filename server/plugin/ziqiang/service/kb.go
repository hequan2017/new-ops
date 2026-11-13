// Package service ziqiang：模型鉴权知识库（需求 3.5-7）
// 口径：法规条目（中美欧三地白名单）+ 判例条目；检索=查询 bigram 对标题+内容的
// 命中计数降序（复用 text 共享 bigram 口径），纯函数可测；种子 15 法规 + 6 判例幂等。
package service

import (
	"sort"
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/text"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// 错误码（ziqiang 段 2700-2799）
const (
	ErrCodeKbParam    = 2712
	ErrCodeKbNotFound = 2713
)

// Jurisdictions 辖区白名单
var Jurisdictions = []string{"中国", "美国", "欧盟"}

func validJurisdiction(j string) bool {
	for _, v := range Jurisdictions {
		if v == j {
			return true
		}
	}
	return false
}

// regHit 检索命中项
type regHit struct {
	model.ZqRegulation
	Score int `json:"score"`
}

// rankRegulations bigram 命中降序（纯函数，测试锚点）：
// 计分/过滤（score>0）/降序排序收敛到 text.RankBigrams（topK 不限量，同分按扫描序
// 稳定排序——扫描集已按 id 升序），再映射回段内 regHit 形状（条目嵌入+score 平铺，
// 保持既有 API 响应结构不变）。
func rankRegulations(list []model.ZqRegulation, query string) []regHit {
	hits := text.RankBigrams(list, func(r model.ZqRegulation) string {
		return r.Title + r.Content
	}, query, 0)
	out := make([]regHit, 0, len(hits))
	for _, h := range hits {
		out = append(out, regHit{ZqRegulation: h.Item, Score: h.Score})
	}
	return out
}

// SearchRegulations 法规检索（辖区过滤 + bigram 打分）
func (s *ZiqiangService) SearchRegulations(query, jurisdiction string, topK int) ([]regHit, error) {
	if strings.TrimSpace(query) == "" {
		return nil, zqErr(ErrCodeKbParam, "检索词必填")
	}
	if jurisdiction != "" && !validJurisdiction(jurisdiction) {
		return nil, zqErr(ErrCodeKbParam, "辖区须为 中国/美国/欧盟")
	}
	if topK < 1 || topK > 50 {
		topK = 5
	}
	list, err := s.repo.RegulationScan(s.dbh(), jurisdiction)
	if err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "扫描法规失败", err)
	}
	hits := rankRegulations(list, query)
	if len(hits) > topK {
		hits = hits[:topK]
	}
	return hits, nil
}

// SaveRegulation 新建/更新法规
func (s *ZiqiangService) SaveRegulation(in *model.ZqRegulation) (*model.ZqRegulation, error) {
	if strings.TrimSpace(in.Title) == "" || !validJurisdiction(in.Jurisdiction) {
		return nil, zqErr(ErrCodeKbParam, "法规名必填且辖区须在白名单内")
	}
	if in.ID > 0 {
		exist, err := s.repo.FindRegulation(s.dbh(), in.ID)
		if err != nil {
			return nil, zqErr(ErrCodeKbNotFound, "法规不存在")
		}
		if err := s.repo.UpdateRegulationColumns(s.dbh(), exist.ID, map[string]any{
			"jurisdiction": in.Jurisdiction, "category": in.Category, "title": in.Title,
			"content": in.Content, "effective_date": in.EffectiveDate,
		}); err != nil {
			return exist, errx.Wrap(ErrCodeZqStore, "更新法规失败", err)
		}
		return exist, nil
	}
	if err := s.repo.CreateRegulation(s.dbh(), in); err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "落库法规失败", err)
	}
	return in, nil
}

// DeleteRegulation 删除法规（软删除）
func (s *ZiqiangService) DeleteRegulation(id uint) error {
	rows, err := s.repo.DeleteRegulation(s.dbh(), id)
	if err != nil {
		return errx.Wrap(ErrCodeZqStore, "删除法规失败", err)
	}
	if rows == 0 {
		return zqErr(ErrCodeKbNotFound, "法规不存在")
	}
	return nil
}

// GetRegulations 法规列表（辖区/分类过滤，倒序分页）
func (s *ZiqiangService) GetRegulations(jurisdiction, category string, page, pageSize int) ([]model.ZqRegulation, int64, error) {
	if jurisdiction != "" && !validJurisdiction(jurisdiction) {
		return nil, 0, zqErr(ErrCodeKbParam, "辖区须为 中国/美国/欧盟")
	}
	list, total, err := s.repo.RegulationList(s.dbh(), jurisdiction, category, page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeZqStore, "查询法规列表失败", err)
	}
	return list, total, nil
}

// SaveCase 新建判例
func (s *ZiqiangService) SaveCase(in *model.ZqCase) (*model.ZqCase, error) {
	if strings.TrimSpace(in.Name) == "" || !validJurisdiction(in.Jurisdiction) {
		return nil, zqErr(ErrCodeKbParam, "案件名必填且辖区须在白名单内")
	}
	if err := s.repo.CreateCase(s.dbh(), in); err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "落库判例失败", err)
	}
	return in, nil
}

// CaseHit 判例命中项
type CaseHit struct {
	model.ZqCase
	Score int `json:"score"`
}

// rankCases 判例 bigram 命中降序（纯函数，测试锚点）：案名+案情+裁判要点。
// 同分决胜=ID 升序（显式口径，区别于 text.RankBigrams 的扫描序稳定排序），
// 故基于 text.Bigrams/text.Score 本地实现排序。
func rankCases(list []model.ZqCase, query string) []CaseHit {
	qSet := text.Bigrams(query)
	hits := make([]CaseHit, 0, len(list))
	for _, c := range list {
		if score := text.Score(qSet, c.Name+c.Summary+c.Ruling); score > 0 {
			hits = append(hits, CaseHit{ZqCase: c, Score: score})
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		return hits[i].ID < hits[j].ID
	})
	return hits
}

// SearchCases 判例检索（辖区过滤 + bigram 打分，口径同 SearchRegulations）
func (s *ZiqiangService) SearchCases(query, jurisdiction string, topK int) ([]CaseHit, error) {
	if strings.TrimSpace(query) == "" {
		return nil, zqErr(ErrCodeKbParam, "检索词必填")
	}
	if jurisdiction != "" && !validJurisdiction(jurisdiction) {
		return nil, zqErr(ErrCodeKbParam, "辖区须为 中国/美国/欧盟")
	}
	if topK < 1 || topK > 50 {
		topK = 5
	}
	list, err := s.repo.CaseScan(s.dbh(), jurisdiction)
	if err != nil {
		return nil, errx.Wrap(ErrCodeZqStore, "扫描判例失败", err)
	}
	hits := rankCases(list, query)
	if len(hits) > topK {
		hits = hits[:topK]
	}
	return hits, nil
}

// GetCases 判例列表（辖区过滤，倒序分页）
func (s *ZiqiangService) GetCases(jurisdiction string, page, pageSize int) ([]model.ZqCase, int64, error) {
	if jurisdiction != "" && !validJurisdiction(jurisdiction) {
		return nil, 0, zqErr(ErrCodeKbParam, "辖区须为 中国/美国/欧盟")
	}
	list, total, err := s.repo.CaseList(s.dbh(), jurisdiction, page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeZqStore, "查询判例列表失败", err)
	}
	return list, total, nil
}
