// Package service fanhua：海派文化知识库（需求 3.4-15）
// 口径：语音库条目（沪剧台词/弄堂俚语/经典朗读 × 男声/女声/童声）；
// 组合检索=语音条目打分（查询分词 token 命中计数，复用 N05 分词口径）+ 平行语料
// 编辑相似度召回（复用 N03 口径），双路合并返回。
package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodeVoiceParam   = 2617
	ErrCodeVoiceNotFound = 2618
)

// VoiceCategories 条目分类白名单
var VoiceCategories = []string{"沪剧台词", "弄堂俚语", "经典朗读"}

// SpeakerTypes 声部白名单
var SpeakerTypes = []string{"男声", "女声", "童声"}

func validIn(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// voiceHit 语音命中项
type voiceHit struct {
	model.FanhuaVoiceEntry
	Score int `json:"score"`
}

// rankVoice 语音条目打分（纯函数，测试锚点）：查询分词 token 在标题+内容中的命中计数
func rankVoice(entries []model.FanhuaVoiceEntry, query string) []voiceHit {
	seen := map[string]bool{}
	var tokens []string
	for _, t := range Segment(query) {
		if !seen[t.Word] {
			seen[t.Word] = true
			tokens = append(tokens, t.Word)
		}
	}
	hits := make([]voiceHit, 0, len(entries))
	for _, e := range entries {
		hay := e.Title + e.Content
		score := 0
		for _, tk := range tokens {
			score += strings.Count(hay, tk)
		}
		if score > 0 {
			hits = append(hits, voiceHit{FanhuaVoiceEntry: e, Score: score})
		}
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	return hits
}

// SearchCulture 组合检索（语音条目 + 平行语料双路）
func (s *FanhuaService) SearchCulture(query, category, speakerType string, topK int) (map[string]any, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("[%d]检索词必填", ErrCodeVoiceParam)
	}
	if category != "" && !validIn(VoiceCategories, category) {
		return nil, fmt.Errorf("[%d]不支持的分类: %s", ErrCodeVoiceParam, category)
	}
	if speakerType != "" && !validIn(SpeakerTypes, speakerType) {
		return nil, fmt.Errorf("[%d]不支持的声部: %s", ErrCodeVoiceParam, speakerType)
	}
	if topK < 1 || topK > 50 {
		topK = 5
	}
	db := global.GVA_DB.Model(&model.FanhuaVoiceEntry{}).Limit(2000)
	if category != "" {
		db = db.Where("category = ?", category)
	}
	if speakerType != "" {
		db = db.Where("speaker_type = ?", speakerType)
	}
	var entries []model.FanhuaVoiceEntry
	if err := db.Find(&entries).Error; err != nil {
		return nil, err
	}
	voiceHits := rankVoice(entries, query)
	if len(voiceHits) > topK {
		voiceHits = voiceHits[:topK]
	}
	paraHits, err := s.SearchParallel(query, topK)
	if err != nil {
		return nil, err
	}
	return map[string]any{"voiceHits": voiceHits, "parallelHits": paraHits}, nil
}

// SaveVoiceEntry 新建/更新语音条目
func (s *FanhuaService) SaveVoiceEntry(in *model.FanhuaVoiceEntry) (*model.FanhuaVoiceEntry, error) {
	if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Content) == "" {
		return nil, fmt.Errorf("[%d]标题与内容必填", ErrCodeVoiceParam)
	}
	if !validIn(VoiceCategories, in.Category) {
		return nil, fmt.Errorf("[%d]分类须在白名单内", ErrCodeVoiceParam)
	}
	if !validIn(SpeakerTypes, in.SpeakerType) {
		in.SpeakerType = "女声"
	}
	if in.ID > 0 {
		var exist model.FanhuaVoiceEntry
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]语音条目不存在", ErrCodeVoiceNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"title": in.Title, "category": in.Category, "content": in.Content,
			"speaker_type": in.SpeakerType, "duration_sec": in.DurationSec,
			"media_path": in.MediaPath, "text_source": in.TextSource,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteVoiceEntry 删除语音条目（软删除）
func (s *FanhuaService) DeleteVoiceEntry(id uint) error {
	res := global.GVA_DB.Delete(&model.FanhuaVoiceEntry{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]语音条目不存在", ErrCodeVoiceNotFound)
	}
	return nil
}

// GetVoiceEntries 语音条目列表（分类/声部过滤，倒序分页）
func (s *FanhuaService) GetVoiceEntries(category, speakerType string, page, pageSize int) ([]model.FanhuaVoiceEntry, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.FanhuaVoiceEntry{})
	if category != "" {
		if !validIn(VoiceCategories, category) {
			return nil, 0, fmt.Errorf("[%d]不支持的分类: %s", ErrCodeVoiceParam, category)
		}
		db = db.Where("category = ?", category)
	}
	if speakerType != "" {
		db = db.Where("speaker_type = ?", speakerType)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaVoiceEntry
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}
