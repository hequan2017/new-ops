// Package service ziqiang：模型鉴权知识库（需求 3.5-7）
// 口径：法规条目（中美欧三地白名单）+ 判例条目；检索=查询 bigram 对标题+内容的
// 命中计数降序（复用 O05 bigramSet），纯函数可测；种子 15 法规 + 6 判例幂等。
package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// 错误码（ziqiang 段 2700-2799）
const (
	ErrCodeKbParam   = 2712
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

// rankRegulations bigram 命中降序（纯函数，测试锚点）
func rankRegulations(list []model.ZqRegulation, query string) []regHit {
	qSet := bigramSet(query)
	hits := make([]regHit, 0, len(list))
	for _, r := range list {
		rSet := bigramSet(r.Title + r.Content)
		score := 0
		for bg := range qSet {
			if rSet[bg] {
				score++
			}
		}
		if score > 0 {
			hits = append(hits, regHit{ZqRegulation: r, Score: score})
		}
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	return hits
}

// SearchRegulations 法规检索（辖区过滤 + bigram 打分）
func (s *ZiqiangService) SearchRegulations(query, jurisdiction string, topK int) ([]regHit, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("[%d]检索词必填", ErrCodeKbParam)
	}
	if jurisdiction != "" && !validJurisdiction(jurisdiction) {
		return nil, fmt.Errorf("[%d]辖区须为 中国/美国/欧盟", ErrCodeKbParam)
	}
	if topK < 1 || topK > 50 {
		topK = 5
	}
	db := global.GVA_DB.Model(&model.ZqRegulation{})
	if jurisdiction != "" {
		db = db.Where("jurisdiction = ?", jurisdiction)
	}
	var list []model.ZqRegulation
	if err := db.Limit(2000).Find(&list).Error; err != nil {
		return nil, err
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
		return nil, fmt.Errorf("[%d]法规名必填且辖区须在白名单内", ErrCodeKbParam)
	}
	if in.ID > 0 {
		var exist model.ZqRegulation
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]法规不存在", ErrCodeKbNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"jurisdiction": in.Jurisdiction, "category": in.Category, "title": in.Title,
			"content": in.Content, "effective_date": in.EffectiveDate,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteRegulation 删除法规（软删除）
func (s *ZiqiangService) DeleteRegulation(id uint) error {
	res := global.GVA_DB.Delete(&model.ZqRegulation{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]法规不存在", ErrCodeKbNotFound)
	}
	return nil
}

// GetRegulations 法规列表（辖区/分类过滤，倒序分页）
func (s *ZiqiangService) GetRegulations(jurisdiction, category string, page, pageSize int) ([]model.ZqRegulation, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.ZqRegulation{})
	if jurisdiction != "" {
		if !validJurisdiction(jurisdiction) {
			return nil, 0, fmt.Errorf("[%d]辖区须为 中国/美国/欧盟", ErrCodeKbParam)
		}
		db = db.Where("jurisdiction = ?", jurisdiction)
	}
	if category != "" {
		db = db.Where("category = ?", category)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ZqRegulation
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// SaveCase 新建判例
func (s *ZiqiangService) SaveCase(in *model.ZqCase) (*model.ZqCase, error) {
	if strings.TrimSpace(in.Name) == "" || !validJurisdiction(in.Jurisdiction) {
		return nil, fmt.Errorf("[%d]案件名必填且辖区须在白名单内", ErrCodeKbParam)
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetCases 判例列表（辖区过滤，倒序分页）
func (s *ZiqiangService) GetCases(jurisdiction string, page, pageSize int) ([]model.ZqCase, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.ZqCase{})
	if jurisdiction != "" {
		if !validJurisdiction(jurisdiction) {
			return nil, 0, fmt.Errorf("[%d]辖区须为 中国/美国/欧盟", ErrCodeKbParam)
		}
		db = db.Where("jurisdiction = ?", jurisdiction)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ZqCase
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
