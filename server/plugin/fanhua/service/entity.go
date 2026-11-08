// Package service fanhua：海派文化实体库（需求 3.4-10 实体构建/词典匹配标注）
// 口径：实体=分类白名单（地域/文化符号/地标）+ 词条 + 别名表；标注=词典匹配计数
// （词条与别名均计入，纯函数），命中按次数降序；标注结果落 fanhua_entity_hit。
package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodeEntityParam = 2611
	ErrCodeEntityNotFound = 2612
)

// EntityCategories 实体分类白名单
var EntityCategories = []string{"地域", "文化符号", "地标"}

// ValidEntityCategory 分类白名单
func ValidEntityCategory(c string) bool {
	for _, v := range EntityCategories {
		if v == c {
			return true
		}
	}
	return false
}

// entityAliases 解析别名表
func entityAliases(e *model.FanhuaEntity) []string {
	var aliases []string
	_ = json.Unmarshal([]byte(e.AliasJSON), &aliases)
	return aliases
}

// countOccurrences 子串计数（纯函数）
func countOccurrences(text, sub string) int {
	if sub == "" {
		return 0
	}
	return strings.Count(text, sub)
}

// EntityHitTerm 命中项
type EntityHitTerm struct {
	Term     string `json:"term"`
	Category string `json:"category"`
	Count    int    `json:"count"`
}

// entityMatchCount 单实体命中数：词条+别名全部候选区间，按起点排序、长匹配优先去重叠后计数
func entityMatchCount(text string, e *model.FanhuaEntity) int {
	type seg struct{ start, end int }
	var segs []seg
	add := func(w string) {
		if w == "" {
			return
		}
		off := 0
		for {
			idx := strings.Index(text[off:], w)
			if idx < 0 {
				break
			}
			segs = append(segs, seg{off + idx, off + idx + len(w)})
			off += idx + len(w)
		}
	}
	add(e.Term)
	for _, a := range entityAliases(e) {
		add(a)
	}
	sort.Slice(segs, func(i, j int) bool {
		if segs[i].start != segs[j].start {
			return segs[i].start < segs[j].start
		}
		return segs[i].end > segs[j].end
	})
	count := 0
	lastEnd := -1
	for _, sg := range segs {
		if sg.start >= lastEnd {
			count++
			lastEnd = sg.end
		}
	}
	return count
}

// entityMatch 词典匹配（纯函数，测试锚点）：词条与别名去重叠计数，按次数降序
func entityMatch(text string, entities []model.FanhuaEntity) []EntityHitTerm {
	hits := make([]EntityHitTerm, 0, 8)
	for i := range entities {
		if count := entityMatchCount(text, &entities[i]); count > 0 {
			hits = append(hits, EntityHitTerm{Term: entities[i].Term, Category: entities[i].Category, Count: count})
		}
	}
	for i := 1; i < len(hits); i++ {
		for j := i; j > 0 && hits[j].Count > hits[j-1].Count; j-- {
			hits[j], hits[j-1] = hits[j-1], hits[j]
		}
	}
	return hits
}

// AnnotateText 实体标注（匹配→结果落库）
func (s *FanhuaService) AnnotateText(text string, corpusId uint) (map[string]any, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("[%d]标注文本必填", ErrCodeEntityParam)
	}
	var entities []model.FanhuaEntity
	if err := global.GVA_DB.Limit(5000).Find(&entities).Error; err != nil {
		return nil, err
	}
	hits := entityMatch(text, entities)
	b, _ := json.Marshal(hits)
	rec := &model.FanhuaEntityHit{CorpusId: corpusId, Text: text, HitsJSON: string(b)}
	if err := global.GVA_DB.Create(rec).Error; err != nil {
		return nil, err
	}
	return map[string]any{"record": rec, "hits": hits}, nil
}

// SaveEntity 新建/更新实体
func (s *FanhuaService) SaveEntity(in *model.FanhuaEntity) (*model.FanhuaEntity, error) {
	if strings.TrimSpace(in.Term) == "" || !ValidEntityCategory(in.Category) {
		return nil, fmt.Errorf("[%d]词条必填且分类须在白名单内", ErrCodeEntityParam)
	}
	if strings.TrimSpace(in.AliasJSON) != "" {
		var aliases []string
		if err := json.Unmarshal([]byte(in.AliasJSON), &aliases); err != nil {
			return nil, fmt.Errorf("[%d]别名须为字符串 JSON 数组", ErrCodeEntityParam)
		}
	}
	if in.ID > 0 {
		var exist model.FanhuaEntity
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]实体不存在", ErrCodeEntityNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"category": in.Category, "term": in.Term, "alias_json": in.AliasJSON, "desc": in.Desc,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteEntity 删除实体（软删除）
func (s *FanhuaService) DeleteEntity(id uint) error {
	res := global.GVA_DB.Delete(&model.FanhuaEntity{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]实体不存在", ErrCodeEntityNotFound)
	}
	return nil
}

// GetEntities 实体列表（分类过滤，倒序分页）
func (s *FanhuaService) GetEntities(category string, page, pageSize int) ([]model.FanhuaEntity, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.FanhuaEntity{})
	if category != "" {
		if !ValidEntityCategory(category) {
			return nil, 0, fmt.Errorf("[%d]不支持的分类: %s", ErrCodeEntityParam, category)
		}
		db = db.Where("category = ?", category)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaEntity
	err := db.Order("category, id").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetEntityHits 标注记录列表（倒序分页）
func (s *FanhuaService) GetEntityHits(page, pageSize int) ([]model.FanhuaEntityHit, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.FanhuaEntityHit{})
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaEntityHit
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
