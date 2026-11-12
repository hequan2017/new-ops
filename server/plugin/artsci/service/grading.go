// Package service artsci：智能作业批改与反馈（需求 3.3-8）
// 口径：ModelId>0 走 llmops 网关 chat 结构化评分（JSON 校验失败重试 1 次），
// ModelId=0 走规则口径 fallback（字数分+评分标准要点覆盖）；chatFn 注入可测。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hequan2017/new-ops/server/plugin/artsci/model"
	llmopsService "github.com/hequan2017/new-ops/server/plugin/llmops/service"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（artsci 段 2500-2599）
const (
	ErrCodeGradeText   = 2510
	ErrCodeGradeLLM    = 2511
	ErrCodeGradeEngine = 2512
)

// gradingChatFn 网关调用点（测试替换锚点）
var gradingChatFn = func(ctx context.Context, modelId uint, prompt string) (string, error) {
	content, _, err := llmopsService.LlmopsService.ChatCompletion(ctx, modelId,
		[]llmopsService.ChatMessage{{Role: "user", Content: prompt}}, nil)
	return content, err
}

// buildGradePrompt 批改提示词（要求只输出 JSON）
func buildGradePrompt(rubric, studentText string) string {
	return "你是作业批改助手。依据评分标准批改学生作答，只输出 JSON，不要任何多余文本：\n" +
		"{\"score\": 0到100的整数, \"comment\": \"总评（80字内）\", \"suggestions\": [\"改进建议1\", \"改进建议2\"]}\n" +
		"评分标准：" + rubric + "\n学生作答：" + studentText
}

// parseGradingJSON 解析批改输出（截取首尾大括号，score∈[0,100] 校验）
func parseGradingJSON(content string) (score float64, comment string, suggestions []string, err error) {
	l := strings.Index(content, "{")
	r := strings.LastIndex(content, "}")
	if l < 0 || r <= l {
		return 0, "", nil, fmt.Errorf("输出不含 JSON 对象")
	}
	var out struct {
		Score       float64  `json:"score"`
		Comment     string   `json:"comment"`
		Suggestions []string `json:"suggestions"`
	}
	if err := json.Unmarshal([]byte(content[l:r+1]), &out); err != nil {
		return 0, "", nil, err
	}
	if out.Score < 0 || out.Score > 100 {
		return 0, "", nil, fmt.Errorf("score 越界: %v", out.Score)
	}
	return out.Score, out.Comment, out.Suggestions, nil
}

// rubricKeywords 评分标准要点提取（按分隔符切词，取 ≥2 字符的词元，至多 8 个）
func rubricKeywords(rubric string) []string {
	fields := strings.FieldsFunc(rubric, func(r rune) bool {
		return r == '，' || r == ',' || r == '、' || r == '；' || r == ';' || r == ' ' || r == '\n' || r == '。'
	})
	seen := map[string]bool{}
	out := make([]string, 0, 8)
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if utf8.RuneCountInString(f) >= 2 && !seen[f] {
			seen[f] = true
			out = append(out, f)
			if len(out) >= 8 {
				break
			}
		}
	}
	return out
}

// ruleGrade 规则口径打分（纯函数，测试锚点）：60 基础 + 字数分(≤20) + 要点覆盖分(≤20)
func ruleGrade(rubric, studentText string) (score float64, comment string, suggestions []string) {
	n := utf8.RuneCountInString(strings.TrimSpace(studentText))
	lenScore := float64(n) / 50.0
	if lenScore > 20 {
		lenScore = 20
	}
	kws := rubricKeywords(rubric)
	hit, miss := 0, make([]string, 0, len(kws))
	for _, k := range kws {
		if strings.Contains(studentText, k) {
			hit++
		} else {
			miss = append(miss, k)
		}
	}
	covScore := 0.0
	if len(kws) > 0 {
		covScore = 20 * float64(hit) / float64(len(kws))
	}
	score = 60 + lenScore + covScore
	comment = fmt.Sprintf("规则口径批改：字数 %d，要点覆盖 %d/%d", n, hit, len(kws))
	if len(miss) > 0 {
		suggestions = append(suggestions, "建议补充要点："+strings.Join(miss, "、"))
	}
	if n < 100 {
		suggestions = append(suggestions, "篇幅偏短，建议扩充论述")
	}
	return score, comment, suggestions
}

// gradeEngine 单条批改（无 DB，测试锚点）：按引擎写回 rec 批改字段
func gradeEngine(rec *model.ArtsciGrading, chatFn func(ctx context.Context, modelId uint, prompt string) (string, error)) error {
	if strings.TrimSpace(rec.StudentText) == "" || strings.TrimSpace(rec.Rubric) == "" {
		return artErr(ErrCodeGradeText, "评分标准与学生作答必填")
	}
	if rec.ModelId == 0 {
		score, comment, sugg := ruleGrade(rec.Rubric, rec.StudentText)
		rec.Engine, rec.Ok, rec.ErrMsg = "rule", true, ""
		rec.Score = round6(score)
		rec.Comment = comment
		b, _ := json.Marshal(sugg)
		rec.Suggestions = string(b)
		return nil
	}
	prompt := buildGradePrompt(rec.Rubric, rec.StudentText)
	var score float64
	var comment string
	var sugg []string
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ { // JSON 校验失败重试 1 次
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		content, err := chatFn(ctx, rec.ModelId, prompt)
		cancel()
		if err != nil {
			lastErr = artWrapErr(ErrCodeGradeLLM, "模型调用失败", err)
			continue
		}
		score, comment, sugg, lastErr = func() (float64, string, []string, error) {
			s, c, sg, e := parseGradingJSON(content)
			return s, c, sg, e
		}()
		if lastErr == nil {
			rec.Engine, rec.Ok, rec.ErrMsg = "llm", true, ""
			rec.Score = round6(score)
			rec.Comment = comment
			b, _ := json.Marshal(sugg)
			rec.Suggestions = string(b)
			return nil
		}
		lastErr = artWrapErr(ErrCodeGradeLLM, "批改输出解析失败", lastErr)
	}
	return lastErr
}

// GradeEssay 批改并落库（成败均落记录）
func (s *ArtsciService) GradeEssay(rec *model.ArtsciGrading) (*model.ArtsciGrading, error) {
	if err := gradeEngine(rec, gradingChatFn); err != nil {
		rec.Ok = false
		rec.ErrMsg = truncateRunText(err.Error())
		if rec.ID == 0 {
			_ = s.repo.CreateGrading(s.dbh(), rec).Error
		} else {
			_ = s.repo.UpdateGradingColumns(s.dbh(), rec.ID,
				map[string]any{"ok": false, "err_msg": rec.ErrMsg}).Error
		}
		return rec, err
	}
	if rec.ID > 0 {
		if err := s.repo.UpdateGradingColumns(s.dbh(), rec.ID, map[string]any{
			"score": rec.Score, "comment": rec.Comment, "suggestions": rec.Suggestions,
			"engine": rec.Engine, "ok": true, "err_msg": "",
		}); err != nil {
			return rec, errx.Wrap(ErrCodeArtsciStore, "更新批改记录失败", err)
		}
		return rec, nil
	}
	if err := s.repo.CreateGrading(s.dbh(), rec); err != nil {
		return rec, errx.Wrap(ErrCodeArtsciStore, "落库批改记录失败", err)
	}
	return rec, nil
}

// GetGradings 批改记录列表（倒序分页）
func (s *ArtsciService) GetGradings(page, pageSize int) ([]model.ArtsciGrading, int64, error) {
	list, total, err := s.repo.GradingList(s.dbh(), page, pageSize)
	if err != nil {
		return nil, 0, errx.Wrap(ErrCodeArtsciStore, "查询批改记录失败", err)
	}
	return list, total, nil
}
