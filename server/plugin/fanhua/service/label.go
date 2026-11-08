// Package service fanhua：文本标注任务（需求 3.4-11/12）
// 口径：train/test 按比例确定性划分（Knuth 散列 id→桶，可复现）；
// 情感规则预标（正/负词表命中计数，多者胜、平局中立）；人工改标覆盖并置 Labeled。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodeLabelParam   = 2613
	ErrCodeLabelNotFound = 2614
)

// posWords 正面词表（含沪语词）
var posWords = []string{"开心", "满意", "漂亮", "好看", "好吃", "高兴", "舒服", "适意", "老灵", "灵额", "赞", "惬意", "开心得来", "交关好", "推板勿多"}

// negWords 负面词表
var negWords = []string{"糟糕", "难吃", "难看", "讨厌", "差劲", "气死", "失望", "后悔", "难听", "恶心", "郁闷", "垃圾", "推板", "戳气", "惹气"}

// sentimentRule 情感规则预标（纯函数，测试锚点）
func sentimentRule(text string) string {
	pos, neg := 0, 0
	for _, w := range posWords {
		pos += strings.Count(text, w)
	}
	for _, w := range negWords {
		neg += strings.Count(text, w)
	}
	switch {
	case neg > pos:
		return "负面"
	case pos > neg:
		return "正面"
	default:
		return "中立"
	}
}

// assignSplit train/test 确定性划分（Knuth 散列，纯函数）
func assignSplit(id uint, trainRatio float64) string {
	bucket := (uint64(id) * 2654435761) % 10000
	if float64(bucket) < trainRatio*10000 {
		return "train"
	}
	return "test"
}

// validateRatio 比例白名单
func validateRatio(r float64) error {
	if r < 0.1 || r > 0.9 {
		return fmt.Errorf("[%d]trainRatio 须在 [0.1,0.9] 内", ErrCodeLabelParam)
	}
	return nil
}

// BuildTextLabels 从语料库构建标注（limit 条，预标+划分）
func (s *FanhuaService) BuildTextLabels(trainRatio float64, limit int) (map[string]any, error) {
	if err := validateRatio(trainRatio); err != nil {
		return nil, err
	}
	if limit < 1 || limit > 1000 {
		limit = 100
	}
	var corpora []model.FanhuaCorpus
	if err := global.GVA_DB.Order("id").Limit(limit).Find(&corpora).Error; err != nil {
		return nil, err
	}
	if len(corpora) == 0 {
		return nil, fmt.Errorf("[%d]语料库为空，无可标注数据", ErrCodeLabelParam)
	}
	labels := make([]model.FanhuaTextLabel, 0, len(corpora))
	for _, c := range corpora {
		labels = append(labels, model.FanhuaTextLabel{
			CorpusId:      c.ID,
			Text:          c.Content,
			Split:         assignSplit(c.ID, trainRatio),
			AutoSentiment: sentimentRule(c.Content),
		})
	}
	if err := global.GVA_DB.CreateInBatches(labels, 200).Error; err != nil {
		return nil, err
	}
	return map[string]any{"created": len(labels)}, nil
}

// UpdateLabel 人工改标（情感/划分覆盖，置 Labeled）
func (s *FanhuaService) UpdateLabel(id uint, sentiment, split string) (*model.FanhuaTextLabel, error) {
	if sentiment != "正面" && sentiment != "中立" && sentiment != "负面" {
		return nil, fmt.Errorf("[%d]情感标签须为 正面/中立/负面", ErrCodeLabelParam)
	}
	if split != "train" && split != "test" && split != "none" {
		return nil, fmt.Errorf("[%d]划分须为 train/test/none", ErrCodeLabelParam)
	}
	var label model.FanhuaTextLabel
	if err := global.GVA_DB.First(&label, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]标注记录不存在", ErrCodeLabelNotFound)
	}
	err := global.GVA_DB.Model(&label).Updates(map[string]any{
		"sentiment": sentiment, "split": split, "labeled": true,
	}).Error
	return &label, err
}

// GetTextLabels 标注列表（划分/情感过滤，倒序分页）
func (s *FanhuaService) GetTextLabels(split, sentiment string, page, pageSize int) ([]model.FanhuaTextLabel, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.FanhuaTextLabel{})
	if split != "" {
		db = db.Where("split = ?", split)
	}
	if sentiment != "" {
		db = db.Where("sentiment = ? OR (sentiment = '' AND auto_sentiment = ?)", sentiment, sentiment)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaTextLabel
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// LabelStats 标注分布统计（划分×情感 聚合）
func (s *FanhuaService) LabelStats() (map[string]any, error) {
	type row struct {
		Split     string `json:"split"`
		Sentiment string `json:"sentiment"`
		Count     int64  `json:"count"`
	}
	var rows []row
	err := global.GVA_DB.Model(&model.FanhuaTextLabel{}).
		Select("split, COALESCE(NULLIF(sentiment, ''), auto_sentiment) as sentiment, COUNT(*) as count").
		Group("split, sentiment").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return map[string]any{"distribution": rows}, nil
}
