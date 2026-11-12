// Package service fanhua：价值观知识库（需求 3.4-16）
// 口径：敏感词 block 拦截/replace 替换两动作；健康度=100-Σ(权重×命中次数) 下限 0；
// 情感三分类复用 N07 词表口径（正/中/负）；词表/模板/冲突案例三类种子幂等（seed.go）。
package service

import (
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodeValuesParam    = 2619
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
		return nil, errx.New(ErrCodeValuesParam, "评估文本必填")
	}
	words, err := s.repo.EnabledWords(s.dbh())
	if err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "查询敏感词失败", err)
	}
	out := computeHealth(text, words)
	return &out, nil
}

// SaveWord 新建/更新敏感词
func (s *FanhuaService) SaveWord(in *model.FanhuaSensitiveWord) (*model.FanhuaSensitiveWord, error) {
	if strings.TrimSpace(in.Word) == "" {
		return nil, errx.New(ErrCodeValuesParam, "敏感词必填")
	}
	if in.Action != "block" && in.Action != "replace" {
		in.Action = "block"
	}
	if in.Weight < 1 || in.Weight > 100 {
		in.Weight = 10
	}
	db := s.dbh()
	if in.ID > 0 {
		exist, err := s.repo.FindSensitiveWord(db, in.ID)
		if err != nil {
			return nil, errx.New(ErrCodeValuesNotFound, "敏感词不存在")
		}
		if err := s.repo.UpdateSensitiveWordColumns(db, exist.ID, map[string]any{
			"word": in.Word, "action": in.Action, "replacement": in.Replacement,
			"weight": in.Weight, "enabled": in.Enabled,
		}); err != nil {
			return nil, errx.Wrap(ErrCodeFanhuaStore, "更新敏感词失败", err)
		}
		return exist, nil
	}
	if err := s.repo.CreateSensitiveWord(db, in); err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "创建敏感词失败", err)
	}
	return in, nil
}

// DeleteWord 删除敏感词（软删除）
func (s *FanhuaService) DeleteWord(id uint) error {
	rows, err := s.repo.DeleteSensitiveWord(s.dbh(), id)
	if err != nil {
		return errx.Wrap(ErrCodeFanhuaStore, "删除敏感词失败", err)
	}
	if rows == 0 {
		return errx.New(ErrCodeValuesNotFound, "敏感词不存在")
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
	if action != "" && action != "block" && action != "replace" {
		return nil, 0, errx.New(ErrCodeValuesParam, "动作须为 block/replace")
	}
	list, total, err := s.repo.WordList(s.dbh(), action, page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeFanhuaStore, "查询敏感词失败", err)
	}
	return list, total, nil
}

// GetTemplates 合规表达模板列表
func (s *FanhuaService) GetTemplates() ([]model.FanhuaComplianceTemplate, error) {
	list, err := s.repo.ComplianceTemplatesAll(s.dbh())
	if err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "查询合规模板失败", err)
	}
	return list, nil
}

// GetConflictCases 价值观冲突案例列表
func (s *FanhuaService) GetConflictCases(lang string) ([]model.FanhuaConflictCase, error) {
	list, err := s.repo.ConflictCasesByLang(s.dbh(), lang)
	if err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "查询冲突案例失败", err)
	}
	return list, nil
}
