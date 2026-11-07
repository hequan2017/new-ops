// Package service exam：题库管理（采购件③）——CRUD + JSONL 批量导入
package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/exam/model"
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
		return fmt.Errorf("[%d]不支持的题型: %s", ErrCodeQTypeBad, q.Type)
	}
	if strings.TrimSpace(q.Stem) == "" {
		return fmt.Errorf("[%d]题干不能为空", ErrCodeQStem)
	}
	if NeedsOptions(q.Type) {
		var opts []string
		if err := json.Unmarshal([]byte(q.OptionsJSON), &opts); err != nil || len(opts) < 2 {
			return fmt.Errorf("[%d]选择类题型须提供 ≥2 个选项", ErrCodeQStem)
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

// ExamService 服务单例
var ExamApp = new(ExamService)

type ExamService struct{}

// GetQuestions 题目列表（题型/知识点过滤，分页）
func (s *ExamService) GetQuestions(page, pageSize int, qType, knowledgePoint string) ([]model.ExamQuestion, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.ExamQuestion{})
	if qType != "" {
		if !ValidQuestionType(qType) {
			return nil, 0, fmt.Errorf("[%d]不支持的题型: %s", ErrCodeQTypeBad, qType)
		}
		db = db.Where("type = ?", qType)
	}
	if knowledgePoint != "" {
		db = db.Where("knowledge_point = ?", knowledgePoint)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.ExamQuestion
	err := db.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// SaveQuestion 新建/更新题目
func (s *ExamService) SaveQuestion(in *model.ExamQuestion) (*model.ExamQuestion, error) {
	if err := validateQuestion(in); err != nil {
		return nil, err
	}
	if in.ID > 0 {
		var q model.ExamQuestion
		if err := global.GVA_DB.First(&q, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]题目不存在", ErrCodeQNotFound)
		}
		err := global.GVA_DB.Model(&q).Updates(map[string]any{
			"type": in.Type, "stem": in.Stem, "options_json": in.OptionsJSON,
			"answer": in.Answer, "score": in.Score, "knowledge_point": in.KnowledgePoint,
			"difficulty": in.Difficulty,
		}).Error
		if err != nil {
			return nil, err
		}
		return &q, nil
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeleteQuestion 删除题目（软删除）
func (s *ExamService) DeleteQuestion(id uint) error {
	res := global.GVA_DB.Delete(&model.ExamQuestion{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]题目不存在", ErrCodeQNotFound)
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
		return nil, skipped, fmt.Errorf("[%d]无有效题目（跳过 %d 行）", ErrCodeQImport, skipped)
	}
	return out, skipped, nil
}

// ImportQuestions JSONL 批量导入
func (s *ExamService) ImportQuestions(content string) (imported, skipped int, err error) {
	questions, skipped, err := ParseQuestionJSONL(content)
	if err != nil {
		return 0, skipped, err
	}
	if err := global.GVA_DB.CreateInBatches(questions, 500).Error; err != nil {
		return 0, skipped, err
	}
	return len(questions), skipped, nil
}
