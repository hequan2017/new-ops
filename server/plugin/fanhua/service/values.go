// Package service fanhua：价值观知识库（需求 3.4-16）
// 口径：敏感词 block 拦截/replace 替换两动作；健康度=100-Σ(权重×命中次数) 下限 0；
// 情感三分类复用 N07 词表口径（正/中/负）；词表/模板/冲突案例三类种子幂等（seed.go）。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodeValuesParam   = 2619
	ErrCodeValuesNotFound = 2620
)

// healthHit 命中项
type healthHit struct {
	Word   string `json:"word"`
	Action string `json:"action"`
	Count  int    `json:"count"`
	Weight int    `json:"weight"`
}

// healthResult 健康度结果
type healthResult struct {
	Health    int         `json:"health"`
	Sentiment string      `json:"sentiment"`
	Blocked   bool        `json:"blocked"`
	Hits      []healthHit `json:"hits"`
	Sanitized string      `json:"sanitized"`
}

// computeHealth 健康度计算（纯函数，测试锚点）：
// 加权扣减 100-Σ(权重×命中) 下限 0；任一 block 词命中即整体拦截；
// replace 词命中输出替换后文本。
func computeHealth(text string, words []model.FanhuaSensitiveWord) healthResult {
	res := healthResult{Health: 100, Sentiment: sentimentRule(text), Hits: []healthHit{}}
	sanitized := text
	for _, w := range words {
		if !w.Enabled {
			continue
		}
		count := strings.Count(text, w.Word)
		if count == 0 {
			continue
		}
		res.Hits = append(res.Hits, healthHit{Word: w.Word, Action: w.Action, Count: count, Weight: w.Weight})
		res.Health -= w.Weight * count
		switch w.Action {
		case "block":
			res.Blocked = true
		case "replace":
			if w.Replacement != "" {
				sanitized = strings.ReplaceAll(sanitized, w.Word, w.Replacement)
			}
		}
	}
	if res.Health < 0 {
		res.Health = 0
	}
	res.Sanitized = sanitized
	return res
}

// GetHealth 健康度评估（读取启用词表）
func (s *FanhuaService) GetHealth(text string) (*healthResult, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("[%d]评估文本必填", ErrCodeValuesParam)
	}
	var words []model.FanhuaSensitiveWord
	if err := global.GVA_DB.Where("enabled = ?", true).Limit(5000).Find(&words).Error; err != nil {
		return nil, err
	}
	out := computeHealth(text, words)
	return &out, nil
}

// SaveWord 新建/更新敏感词
func (s *FanhuaService) SaveWord(in *model.FanhuaSensitiveWord) (*model.FanhuaSensitiveWord, error) {
	if strings.TrimSpace(in.Word) == "" {
		return nil, fmt.Errorf("[%d]敏感词必填", ErrCodeValuesParam)
	}
	if in.Action != "block" && in.Action != "replace" {
		in.Action = "block"
	}
	if in.Weight < 1 || in.Weight > 100 {
		in.Weight = 10
	}
	if in.ID > 0 {
		var exist model.FanhuaSensitiveWord
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]敏感词不存在", ErrCodeValuesNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"word": in.Word, "action": in.Action, "replacement": in.Replacement,
			"weight": in.Weight, "enabled": in.Enabled,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteWord 删除敏感词（软删除）
func (s *FanhuaService) DeleteWord(id uint) error {
	res := global.GVA_DB.Delete(&model.FanhuaSensitiveWord{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]敏感词不存在", ErrCodeValuesNotFound)
	}
	return nil
}

// GetWords 敏感词列表（动作过滤，倒序分页）
func (s *FanhuaService) GetWords(action string, page, pageSize int) ([]model.FanhuaSensitiveWord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.FanhuaSensitiveWord{})
	if action != "" {
		if action != "block" && action != "replace" {
			return nil, 0, fmt.Errorf("[%d]动作须为 block/replace", ErrCodeValuesParam)
		}
		db = db.Where("action = ?", action)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.FanhuaSensitiveWord
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetTemplates 合规表达模板列表
func (s *FanhuaService) GetTemplates() ([]model.FanhuaComplianceTemplate, error) {
	var list []model.FanhuaComplianceTemplate
	err := global.GVA_DB.Order("id").Find(&list).Error
	return list, err
}

// GetConflictCases 价值观冲突案例列表
func (s *FanhuaService) GetConflictCases(lang string) ([]model.FanhuaConflictCase, error) {
	db := global.GVA_DB.Model(&model.FanhuaConflictCase{})
	if lang != "" {
		db = db.Where("lang = ?", lang)
	}
	var list []model.FanhuaConflictCase
	err := db.Order("id").Limit(200).Find(&list).Error
	return list, err
}
