// Package service fanhua：海派文化实体库（需求 3.4-10 实体构建/词典匹配标注）
// 口径：实体=分类白名单（地域/文化符号/地标）+ 词条 + 别名表；标注=词典匹配计数
// （词条与别名均计入，纯函数），命中按次数降序；标注结果落 fanhua_entity_hit。
package service

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodeEntityParam    = 2611
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
		return nil, errx.New(ErrCodeEntityParam, "标注文本必填")
	}
	db := s.dbh()
	entities, err := s.repo.EntitiesScan(db)
	if err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "扫描实体失败", err)
	}
	hits := entityMatch(text, entities)
	b, _ := json.Marshal(hits)
	rec := &model.FanhuaEntityHit{CorpusId: corpusId, Text: text, HitsJSON: string(b)}
	if err := s.repo.CreateEntityHit(db, rec); err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "落标注记录失败", err)
	}
	return map[string]any{"record": rec, "hits": hits}, nil
}

// SaveEntity 新建/更新实体
func (s *FanhuaService) SaveEntity(in *model.FanhuaEntity) (*model.FanhuaEntity, error) {
	if strings.TrimSpace(in.Term) == "" || !ValidEntityCategory(in.Category) {
		return nil, errx.New(ErrCodeEntityParam, "词条必填且分类须在白名单内")
	}
	if strings.TrimSpace(in.AliasJSON) != "" {
		var aliases []string
		if err := json.Unmarshal([]byte(in.AliasJSON), &aliases); err != nil {
			return nil, errx.New(ErrCodeEntityParam, "别名须为字符串 JSON 数组")
		}
	}
	db := s.dbh()
	if in.ID > 0 {
		exist, err := s.repo.FindEntity(db, in.ID)
		if err != nil {
			return nil, errx.New(ErrCodeEntityNotFound, "实体不存在")
		}
		if err := s.repo.UpdateEntityColumns(db, exist.ID, map[string]any{
			"category": in.Category, "term": in.Term, "alias_json": in.AliasJSON, "desc": in.Desc,
		}); err != nil {
			return nil, errx.Wrap(ErrCodeFanhuaStore, "更新实体失败", err)
		}
		return exist, nil
	}
	if err := s.repo.CreateEntity(db, in); err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "创建实体失败", err)
	}
	return in, nil
}

// DeleteEntity 删除实体（软删除）
func (s *FanhuaService) DeleteEntity(id uint) error {
	rows, err := s.repo.DeleteEntity(s.dbh(), id)
	if err != nil {
		return errx.Wrap(ErrCodeFanhuaStore, "删除实体失败", err)
	}
	if rows == 0 {
		return errx.New(ErrCodeEntityNotFound, "实体不存在")
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
	if category != "" && !ValidEntityCategory(category) {
		return nil, 0, errx.New(ErrCodeEntityParam, "不支持的分类: "+category)
	}
	list, total, err := s.repo.EntityList(s.dbh(), category, page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeFanhuaStore, "查询实体失败", err)
	}
	return list, total, nil
}

// GetEntityHits 标注记录列表（倒序分页）
func (s *FanhuaService) GetEntityHits(page, pageSize int) ([]model.FanhuaEntityHit, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	list, total, err := s.repo.EntityHitList(s.dbh(), page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeFanhuaStore, "查询标注记录失败", err)
	}
	return list, total, nil
}
