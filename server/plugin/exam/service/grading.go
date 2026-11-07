// Package service exam：判分引擎（采购件③）——客观自动判分 + 主观/实操 LLM 判分
// 口径：single/judge 归一全等；multiple 全对满分、漏选按比例（可关）、错选零分；
// short/practice 走 LLM（评分标准 prompt，{"score":0-1 比例,"reason"} 容错解析）；
// 判分核心注入 chatFn 可测。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/exam/model"
	llmopsService "github.com/hequan2017/new-ops/server/plugin/llmops/service"
)

// 错误码（exam 段续用）
const (
	ErrCodeGradeState = 2414
	ErrCodeGradeJudge = 2415
)

// normalizeAnswer 答案归一（去空白+小写）
func normalizeAnswer(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// parseAnswerSet 多选答案集合解析（分隔符 , 、 空格）
func parseAnswerSet(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == '、' || r == ' ' || r == ';' || r == '/'
	})
	set := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" {
			set = append(set, strings.ToUpper(f))
		}
	}
	sort.Strings(set)
	return set
}

// expandCompactSet 多选连写扩展（纯分隔符缺失时逐字符拆，如 ABD → A,B,D）
func expandCompactSet(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || strings.ContainsAny(s, ",、 /;") {
		return s
	}
	r := []rune(s)
	if len(r) < 2 {
		return s
	}
	for _, ch := range r { // 非字母数字含中文 → 原样（短答题不走此路径）
		if !(ch >= 'A' && ch <= 'Z') && !(ch >= '0' && ch <= '9') {
			return s
		}
	}
	var sb strings.Builder
	for i, ch := range r {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteRune(ch)
	}
	return sb.String()
}

// GradeObjective 客观题判分（纯函数，测试锚点）
func GradeObjective(qType, standard, answer string, fullScore float64, multiPartial bool) (float64, string) {
	if qType == "single" {
		if normalizeAnswer(answer) == normalizeAnswer(standard) {
			return fullScore, "命中标准答案"
		}
		return 0, "与标准答案不符"
	}
	if qType == "judge" {
		if normalizeJudge(answer) == normalizeJudge(standard) {
			return fullScore, "判断正确"
		}
		return 0, "判断错误"
	}
	// multiple（无分隔符连写如 "ABD" 按字符拆分，兼容 "A,B,D" 两种书写）
	stdSet := parseAnswerSet(expandCompactSet(standard))
	ansSet := parseAnswerSet(expandCompactSet(answer))
	if len(ansSet) == 0 {
		return 0, "未作答"
	}
	stdMap := map[string]bool{}
	for _, s := range stdSet {
		stdMap[s] = true
	}
	hits, wrong := 0, 0
	for _, a := range ansSet {
		if stdMap[a] {
			hits++
		} else {
			wrong++
		}
	}
	if wrong > 0 {
		return 0, fmt.Sprintf("错选 %d 项", wrong)
	}
	if len(ansSet) == len(stdSet) {
		return fullScore, "全对"
	}
	if multiPartial && len(stdSet) > 0 {
		score := fullScore * float64(hits) / float64(len(stdSet))
		return score, fmt.Sprintf("漏选 %d 项，按比例得分", len(stdSet)-hits)
	}
	return 0, "漏选不得分"
}

// normalizeJudge 判断题归一（对/错/true/false/√/×）
func normalizeJudge(s string) string {
	s = normalizeAnswer(s)
	switch s {
	case "对", "正确", "true", "√", "是", "y", "yes":
		return "true"
	case "错", "错误", "false", "×", "否", "n", "no":
		return "false"
	}
	return s
}

// buildGradePrompt 主观题评分 prompt（纯函数）
func buildGradePrompt(stem, standard, answer string) string {
	return "你是阅卷老师。按标准答案给考生回答打分，score 为 0 到 1 的比例。只输出 JSON：\n" +
		`{"score":0.8,"reason":"一句话理由"}` + "\n题目：" + stem + "\n标准答案/评分标准：" + standard + "\n考生回答：" + answer
}

// extractJSONObjectLocal 均衡 JSON 截取（与 llmops 同口径，exam 本地实现）
func extractJSONObject(text string) (string, bool) {
	start := strings.Index(text, "{")
	if start < 0 {
		return "", false
	}
	depth, inStr, escape := 0, false, false
	for i := start; i < len(text); i++ {
		c := text[i]
		if escape {
			escape = false
			continue
		}
		switch c {
		case '\\':
			if inStr {
				escape = true
			}
		case '"':
			inStr = !inStr
		case '{':
			if !inStr {
				depth++
			}
		case '}':
			if !inStr {
				depth--
				if depth == 0 {
					return text[start : i+1], true
				}
			}
		}
	}
	return "", false
}

// ParseGradeScore LLM 判分输出解析（纯函数：均衡截取 + 0-1 夹取）
func ParseGradeScore(text string) (float64, string, error) {
	obj, ok := extractJSONObject(text)
	if !ok {
		return 0, "", fmt.Errorf("[%d]评判输出解析失败", ErrCodeGradeJudge)
	}
	var parsed struct {
		Score  *float64 `json:"score"`
		Reason string   `json:"reason"`
	}
	if err := json.Unmarshal([]byte(obj), &parsed); err != nil || parsed.Score == nil {
		return 0, "", fmt.Errorf("[%d]评判输出解析失败", ErrCodeGradeJudge)
	}
	score := *parsed.Score
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return score, strings.TrimSpace(parsed.Reason), nil
}

// judgeFn 评判模型注入点
type judgeFn func(question, expected, answer string) (float64, string, error)

// gradeQuestions 判分核心（注入式，测试锚点）
func gradeQuestions(qs []model.ExamQuestion, answers map[string]string, multiPartial bool, judge judgeFn) ([]model.ExamGrade, float64) {
	grades := make([]model.ExamGrade, 0, len(qs))
	var total float64
	for _, q := range qs {
		answer := answers[fmt.Sprintf("%d", q.ID)]
		g := model.ExamGrade{QuestionId: q.ID, Type: q.Type, Answer: answer, Standard: q.Answer, FullScore: q.Score}
		switch q.Type {
		case "single", "judge", "multiple":
			g.Method = "auto"
			g.Score, g.Detail = GradeObjective(q.Type, q.Answer, answer, q.Score, multiPartial)
		default: // short/practice
			g.Method = "llm"
			if judge == nil {
				g.Score, g.Detail = 0, "未配置评判模型"
			} else {
				ratio, reason, err := judge(q.Stem, q.Answer, answer)
				if err != nil {
					g.Score, g.Detail = 0, "评判失败: "+err.Error()
				} else {
					g.Score = q.Score * ratio
					g.Detail = reason
				}
			}
		}
		total += g.Score
		grades = append(grades, g)
	}
	return grades, total
}

// GradeSheet 整卷判分（submitted 态前置；结果落库；返回总分）
func (s *ExamService) GradeSheet(sheetId uint, judgeModelId uint, multiPartial bool) (float64, error) {
	var sheet model.ExamAnswerSheet
	if err := global.GVA_DB.First(&sheet, sheetId).Error; err != nil {
		return 0, fmt.Errorf("[%d]答卷不存在", ErrCodeSheetBad)
	}
	if sheet.Status != "submitted" {
		return 0, fmt.Errorf("[%d]答卷未交卷（状态 %s）", ErrCodeGradeState, sheet.Status)
	}
	var exam model.ExamExam
	if err := global.GVA_DB.First(&exam, sheet.ExamId).Error; err != nil {
		return 0, err
	}
	var paper model.ExamPaper
	if err := global.GVA_DB.First(&paper, exam.PaperId).Error; err != nil {
		return 0, err
	}
	questions, _, answers, err := s.GetSheetQuestions(sheetId)
	if err != nil {
		return 0, err
	}
	// 拉题已脱敏，重取含答案版本
	if paper.QuestionIds != "" {
		if err := global.GVA_DB.Where("id IN ?", paperQuestionIds(paper.QuestionIds)).Find(&questions).Error; err != nil {
			return 0, err
		}
	}
	_ = json.Unmarshal([]byte(sheet.AnswersJSON), &answers)

	var judge judgeFn
	if judgeModelId > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		judge = func(question, expected, answer string) (float64, string, error) {
			content, _, err := llmopsService.LlmopsService.ChatCompletion(ctx, judgeModelId,
				[]llmopsService.ChatMessage{{Role: "user", Content: buildGradePrompt(question, expected, answer)}}, nil)
			if err != nil {
				return 0, "", err
			}
			return ParseGradeScore(content)
		}
	}
	grades, total := gradeQuestions(questions, answers, multiPartial, judge)
	if err := global.GVA_DB.Where("sheet_id = ?", sheetId).Delete(&model.ExamGrade{}).Error; err != nil {
		return 0, err
	}
	for i := range grades {
		grades[i].SheetId = sheetId
	}
	if err := global.GVA_DB.CreateInBatches(grades, 200).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// paperQuestionIds 试卷题目 ID 串解析
func paperQuestionIds(idsCSV string) []uint {
	idStrs := strings.Split(idsCSV, ",")
	ids := make([]uint, 0, len(idStrs))
	for _, s := range idStrs {
		if id, err := strconvParseUint(strings.TrimSpace(s)); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func strconvParseUint(s string) (uint, error) {
	var v uint64
	_, err := fmt.Sscanf(s, "%d", &v)
	return uint(v), err
}

// GradeIn 判分请求
type GradeIn struct {
	SheetId      uint `json:"sheetId" binding:"required"`
	JudgeModelId uint `json:"judgeModelId"`
	MultiPartial bool `json:"multiPartial"`
}

// GetExamGrades 判分明细查询
func (s *ExamService) GetExamGrades(sheetId uint) ([]model.ExamGrade, error) {
	var list []model.ExamGrade
	err := global.GVA_DB.Where("sheet_id = ?", sheetId).Order("question_id").Find(&list).Error
	return list, err
}
