// Package service exam：组卷引擎（采购件③）——规则抽题 + 全局去重 + 分值合计
// 口径：规则 {"sections":[{type,count,knowledgePoint?,maxDifficulty?}]}；
// 抽题随机源注入可测；题目跨 section 全局唯一；某 section 题量不足即报错命名该节。
package service

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/exam/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（exam 段续用）
const (
	ErrCodePaperNotFound = 2404
	ErrCodePaperRule     = 2405
	ErrCodePaperLack     = 2406
	paperMaxSections     = 20
)

// PaperSection 抽题规则节
type PaperSection struct {
	Type           string `json:"type"`           // 题型（必填）
	Count          int    `json:"count"`          // 抽题数
	KnowledgePoint string `json:"knowledgePoint"` // 知识点（可选）
	MaxDifficulty  int    `json:"maxDifficulty"`  // 难度上限（可选，1-5）
}

// PaperRule 组卷规则
type PaperRule struct {
	Sections []PaperSection `json:"sections"`
}

// ParsePaperRule 规则解析（纯函数）
func ParsePaperRule(raw string) (*PaperRule, error) {
	var rule PaperRule
	if err := json.Unmarshal([]byte(raw), &rule); err != nil {
		return nil, examWrapErr(ErrCodePaperRule, "组卷规则解析失败", err)
	}
	if len(rule.Sections) == 0 || len(rule.Sections) > paperMaxSections {
		return nil, examErr(ErrCodePaperRule, "规则节数须为 1-%d", paperMaxSections)
	}
	for i := range rule.Sections {
		s := &rule.Sections[i]
		if !ValidQuestionType(s.Type) {
			return nil, examErr(ErrCodePaperRule, "第 %d 节题型非法: %s", i+1, s.Type)
		}
		if s.Count <= 0 {
			return nil, examErr(ErrCodePaperRule, "第 %d 节抽题数须 ≥1", i+1)
		}
		if s.MaxDifficulty < 0 || s.MaxDifficulty > 5 {
			return nil, examErr(ErrCodePaperRule, "第 %d 节难度上限非法", i+1)
		}
	}
	return &rule, nil
}

// PickQuestions 抽题核心（纯函数注入 rand，测试锚点）：返回题目 ID 序列与总分
func PickQuestions(pool []model.ExamQuestion, rule *PaperRule, rng *rand.Rand) ([]uint, float64, error) {
	used := map[uint]bool{}
	var ids []uint
	var total float64
	for si := range rule.Sections {
		sec := &rule.Sections[si]
		var candidates []model.ExamQuestion
		for _, q := range pool {
			if used[q.ID] || q.Type != sec.Type {
				continue
			}
			if sec.KnowledgePoint != "" && q.KnowledgePoint != sec.KnowledgePoint {
				continue
			}
			if sec.MaxDifficulty > 0 && q.Difficulty > sec.MaxDifficulty {
				continue
			}
			candidates = append(candidates, q)
		}
		if len(candidates) < sec.Count {
			return nil, 0, examErr(ErrCodePaperLack, "第 %d 节(%s)题量不足：需 %d 仅 %d", si+1, sec.Type, sec.Count, len(candidates))
		}
		rng.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
		for _, q := range candidates[:sec.Count] {
			used[q.ID] = true
			ids = append(ids, q.ID)
			total += q.Score
		}
	}
	return ids, total, nil
}

// GetPapers 试卷列表
func (s *ExamService) GetPapers() ([]model.ExamPaper, error) {
	list, err := s.repo.PaperList(s.dbh())
	if err != nil {
		return nil, errx.Wrap(ErrCodeExamStore, "查询试卷列表失败", err)
	}
	return list, nil
}

// SavePaper 保存试卷（规则即检）
func (s *ExamService) SavePaper(in *model.ExamPaper) (*model.ExamPaper, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" {
		return nil, examErr(ErrCodePaperRule, "试卷名必填")
	}
	if _, err := ParsePaperRule(in.RuleJSON); err != nil {
		return nil, err
	}
	count, err := s.repo.CountPapersByName(s.dbh(), in.Name, in.ID)
	if err != nil {
		return nil, errx.Wrap(ErrCodeExamStore, "查询试卷重名失败", err)
	}
	if count > 0 {
		return nil, examErr(ErrCodePaperRule, "试卷名已存在")
	}
	if in.ID > 0 {
		p, err := s.repo.FindPaper(s.dbh(), in.ID)
		if err != nil {
			return nil, examErr(ErrCodePaperNotFound, "试卷不存在")
		}
		if err := s.repo.UpdatePaperColumns(s.dbh(), p.ID, map[string]any{
			"name": in.Name, "rule_json": in.RuleJSON, "status": "draft",
		}); err != nil {
			return nil, errx.Wrap(ErrCodeExamStore, "更新试卷失败", err)
		}
		return p, nil
	}
	if in.Status == "" {
		in.Status = "draft"
	}
	if err := s.repo.CreatePaper(s.dbh(), in); err != nil {
		return nil, errx.Wrap(ErrCodeExamStore, "落库试卷失败", err)
	}
	return in, nil
}

// ComposePaper 组卷：按规则抽题，锁定 ID 与总分
func (s *ExamService) ComposePaper(paperId uint) (*model.ExamPaper, error) {
	db := s.dbh()
	p, err := s.repo.FindPaper(db, paperId)
	if err != nil {
		return nil, examErr(ErrCodePaperNotFound, "试卷不存在")
	}
	rule, err := ParsePaperRule(p.RuleJSON)
	if err != nil {
		return nil, err
	}
	pool, err := s.repo.QuestionsAll(db)
	if err != nil {
		return nil, errx.Wrap(ErrCodeExamStore, "查询题库失败", err)
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	ids, total, err := PickQuestions(pool, rule, rng)
	if err != nil {
		return nil, err
	}
	idStrs := make([]string, len(ids))
	for i, id := range ids {
		idStrs[i] = strconv.FormatUint(uint64(id), 10)
	}
	if err := s.repo.UpdatePaperColumns(db, p.ID, map[string]any{
		"question_ids": strings.Join(idStrs, ","), "total_score": total, "status": "ready",
	}); err != nil {
		return nil, errx.Wrap(ErrCodeExamStore, "更新试卷失败", err)
	}
	p.QuestionIds, p.TotalScore, p.Status = strings.Join(idStrs, ","), total, "ready"
	return p, nil
}

// DeletePaper 删除试卷（软删除）
func (s *ExamService) DeletePaper(id uint) error {
	n, err := s.repo.DeletePaper(s.dbh(), id)
	if err != nil {
		return errx.Wrap(ErrCodeExamStore, "删除试卷失败", err)
	}
	if n == 0 {
		return examErr(ErrCodePaperNotFound, "试卷不存在")
	}
	return nil
}
