// Package service fanhua：海派文化知识库（需求 3.4-15）
// 口径：语音库条目（沪剧台词/弄堂俚语/经典朗读 × 男声/女声/童声）；
// 组合检索=语音条目打分（查询 bigram 命中计数，text.Bigrams/text.Score 公共口径）+
// 平行语料编辑相似度召回（复用 N03 口径），双路合并返回。
package service

import (
	"sort"
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/text"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodeVoiceParam    = 2617
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

// rankVoice 语音条目打分（纯函数，测试锚点）：查询 bigram 集对标题+内容命中计数。
// 原实现为分词 token 命中多重计数（非稳定排序），与 text.RankBigrams 的集合计数+稳定排序
// 不等价；按去重要求改用 text.Bigrams/text.Score 公共打分原语，score>0 过滤与本地降序
// 排序保持原口径（同分序由 sort.Slice 决定，与原实现一致）。
func rankVoice(entries []model.FanhuaVoiceEntry, query string) []voiceHit {
	qSet := text.Bigrams(query)
	hits := make([]voiceHit, 0, len(entries))
	for _, e := range entries {
		if score := text.Score(qSet, e.Title+e.Content); score > 0 {
			hits = append(hits, voiceHit{FanhuaVoiceEntry: e, Score: score})
		}
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	return hits
}

// SearchCulture 组合检索（语音条目 + 平行语料双路）
func (s *FanhuaService) SearchCulture(query, category, speakerType string, topK int) (map[string]any, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errx.New(ErrCodeVoiceParam, "检索词必填")
	}
	if category != "" && !validIn(VoiceCategories, category) {
		return nil, errx.New(ErrCodeVoiceParam, "不支持的分类: "+category)
	}
	if speakerType != "" && !validIn(SpeakerTypes, speakerType) {
		return nil, errx.New(ErrCodeVoiceParam, "不支持的声部: "+speakerType)
	}
	if topK < 1 || topK > 50 {
		topK = 5
	}
	entries, err := s.repo.VoiceEntriesScan(s.dbh(), category, speakerType)
	if err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "扫描语音条目失败", err)
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
		return nil, errx.New(ErrCodeVoiceParam, "标题与内容必填")
	}
	if !validIn(VoiceCategories, in.Category) {
		return nil, errx.New(ErrCodeVoiceParam, "分类须在白名单内")
	}
	if !validIn(SpeakerTypes, in.SpeakerType) {
		in.SpeakerType = "女声"
	}
	db := s.dbh()
	if in.ID > 0 {
		exist, err := s.repo.FindVoiceEntry(db, in.ID)
		if err != nil {
			return nil, errx.New(ErrCodeVoiceNotFound, "语音条目不存在")
		}
		if err := s.repo.UpdateVoiceEntryColumns(db, exist.ID, map[string]any{
			"title": in.Title, "category": in.Category, "content": in.Content,
			"speaker_type": in.SpeakerType, "duration_sec": in.DurationSec,
			"media_path": in.MediaPath, "text_source": in.TextSource,
		}); err != nil {
			return nil, errx.Wrap(ErrCodeFanhuaStore, "更新语音条目失败", err)
		}
		return exist, nil
	}
	if err := s.repo.CreateVoiceEntry(db, in); err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "创建语音条目失败", err)
	}
	return in, nil
}

// DeleteVoiceEntry 删除语音条目（软删除）
func (s *FanhuaService) DeleteVoiceEntry(id uint) error {
	rows, err := s.repo.DeleteVoiceEntry(s.dbh(), id)
	if err != nil {
		return errx.Wrap(ErrCodeFanhuaStore, "删除语音条目失败", err)
	}
	if rows == 0 {
		return errx.New(ErrCodeVoiceNotFound, "语音条目不存在")
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
	if category != "" && !validIn(VoiceCategories, category) {
		return nil, 0, errx.New(ErrCodeVoiceParam, "不支持的分类: "+category)
	}
	list, total, err := s.repo.VoiceEntryList(s.dbh(), category, speakerType, page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeFanhuaStore, "查询语音条目失败", err)
	}
	return list, total, nil
}
