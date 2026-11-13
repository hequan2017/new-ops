// Package service exam：题库管理（采购件③）——CRUD + JSONL 批量导入
package service

import (
	"encoding/json"
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/exam/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（exam 段 2400-2499）
const (
	ErrCodeQNotFound = 2400
	ErrCodeQTypeBad  = 2401
	ErrCodeQImport   = 2402
	ErrCodeQStem     = 2403
)

// QuestionTypes 题型白名单
var QuestionTypes = []string{"single", "multiple", "judge", "short", "practice"}

// ValidQuestionType 题型白名单
func ValidQuestionType(t string) bool {
	for _, v := range QuestionTypes {
		if v == t {
			return true
		}
	}
	return false
}

// NeedsOptions 选择类题型（选项必填）
func NeedsOptions(t string) bool {
	return t == "single" || t == "multiple"
}

// validateQuestion 单题校验（导入与保存共用）
func validateQuestion(q *model.ExamQuestion) error {
	if !ValidQuestionType(q.Type) {
		return examErr(ErrCodeQTypeBad, "不支持的题型: %s", q.Type)
	}
	if strings.TrimSpace(q.Stem) == "" {
		return examErr(ErrCodeQStem, "题干不能为空")
	}
	if NeedsOptions(q.Type) {
		var opts []string
		if err := json.Unmarshal([]byte(q.OptionsJSON), &opts); err != nil || len(opts) < 2 {
			return examErr(ErrCodeQStem, "选择类题型须提供 ≥2 个选项")
		}
	}
	if q.Score <= 0 {
		q.Score = 1
	}
	if q.Difficulty <= 0 {
		q.Difficulty = 1
	}
	if q.Difficulty > 5 {
		q.Difficulty = 5
	}
	return nil
}

// ExamService 与 ExamApp 单例定义见 enter.go（依赖注入口径）

// GetQuestions 题目列表（题型/知识点过滤，分页）
// 分页口径：page≤0 回第一页、pageSize 缺省 20 / 上限 100——归一后与 data.Page 等价，
// 归一在 service 完成，repo 侧直接走共享分页。
func (s *ExamService) GetQuestions(page, pageSize int, qType, knowledgePoint string) ([]model.ExamQuestion, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	if qType != "" && !ValidQuestionType(qType) {
		return nil, 0, examErr(ErrCodeQTypeBad, "不支持的题型: %s", qType)
	}
	list, total, err := s.repo.QuestionList(s.dbh(), page, pageSize, qType, knowledgePoint)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeExamStore, "查询题目列表失败", err)
	}
	return list, total, nil
}

// SaveQuestion 新建/更新题目
func (s *ExamService) SaveQuestion(in *model.ExamQuestion) (*model.ExamQuestion, error) {
	if err := validateQuestion(in); err != nil {
		return nil, err
	}
	if in.ID > 0 {
		q, err := s.repo.FindQuestion(s.dbh(), in.ID)
		if err != nil {
			return nil, examErr(ErrCodeQNotFound, "题目不存在")
		}
		if err := s.repo.UpdateQuestionColumns(s.dbh(), q.ID, map[string]any{
			"type": in.Type, "stem": in.Stem, "options_json": in.OptionsJSON,
			"answer": in.Answer, "score": in.Score, "knowledge_point": in.KnowledgePoint,
			"difficulty": in.Difficulty,
		}); err != nil {
			return nil, errx.Wrap(ErrCodeExamStore, "更新题目失败", err)
		}
		return q, nil
	}
	if err := s.repo.CreateQuestion(s.dbh(), in); err != nil {
		return nil, errx.Wrap(ErrCodeExamStore, "落库题目失败", err)
	}
	return in, nil
}

// DeleteQuestion 删除题目（软删除）
func (s *ExamService) DeleteQuestion(id uint) error {
	n, err := s.repo.DeleteQuestion(s.dbh(), id)
	if err != nil {
		return errx.Wrap(ErrCodeExamStore, "删除题目失败", err)
	}
	if n == 0 {
		return examErr(ErrCodeQNotFound, "题目不存在")
	}
	return nil
}

// questionJSONL JSONL 单行结构
type questionJSONL struct {
	Type           string   `json:"type"`
	Stem           string   `json:"stem"`
	Options        []string `json:"options"`
	Answer         string   `json:"answer"`
	Score          float64  `json:"score"`
	KnowledgePoint string   `json:"knowledgePoint"`
	Difficulty     int      `json:"difficulty"`
}

// ParseQuestionJSONL JSONL 解析（纯函数，测试锚点）：逐行校验，坏行跳过计数
func ParseQuestionJSONL(content string) ([]model.ExamQuestion, int, error) {
	var out []model.ExamQuestion
	skipped := 0
	for i, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var row questionJSONL
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			skipped++
			continue
		}
		q := model.ExamQuestion{
			Type: row.Type, Stem: row.Stem, Answer: row.Answer,
			Score: row.Score, KnowledgePoint: row.KnowledgePoint, Difficulty: row.Difficulty,
		}
		if len(row.Options) > 0 {
			b, _ := json.Marshal(row.Options)
			q.OptionsJSON = string(b)
		}
		if err := validateQuestion(&q); err != nil {
			skipped++ // 第 %d 行坏行计数（错误汇总由调用方按需展示）
			_ = i
			continue
		}
		out = append(out, q)
	}
	if len(out) == 0 {
		return nil, skipped, examErr(ErrCodeQImport, "无有效题目（跳过 %d 行）", skipped)
	}
	return out, skipped, nil
}

// ImportQuestions JSONL 批量导入
func (s *ExamService) ImportQuestions(content string) (imported, skipped int, err error) {
	questions, skipped, err := ParseQuestionJSONL(content)
	if err != nil {
		return 0, skipped, err
	}
	if err := s.repo.CreateQuestionsBatch(s.dbh(), questions); err != nil {
		return 0, skipped, errx.Wrap(ErrCodeExamStore, "批量落库题目失败", err)
	}
	return len(questions), skipped, nil
}
