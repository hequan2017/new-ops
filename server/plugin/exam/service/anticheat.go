// Package service exam：防作弊（采购件③）——行为上报留痕 + 主观题答案相似度抽查
// 口径：事件白名单 blur/paste/copy/fullscreen_exit；相似度抽查在同考次已交卷答卷间
// 按题两两比对（Levenshtein 归一，组合数上限 500 防爆炸），≥0.9 判可疑返回 top N。
package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/exam/model"
)

// 错误码（exam 段续用）
const (
	ErrCodeCheatBad   = 2418
	cheatSimThreshold = 0.9
	cheatMaxPairs     = 500
)

// BehaviorEvents 行为事件白名单
var BehaviorEvents = []string{"blur", "paste", "copy", "fullscreen_exit"}

// ValidBehaviorEvent 事件白名单
func ValidBehaviorEvent(e string) bool {
	for _, v := range BehaviorEvents {
		if v == e {
			return true
		}
	}
	return false
}

// editSimilarity 归一编辑相似度（纯函数：1-dist/maxLen，空串对空串=1）
func editSimilarity(a, b string) float64 {
	ra, rb := []rune(strings.TrimSpace(a)), []rune(strings.TrimSpace(b))
	la, lb := len(ra), len(rb)
	if la == 0 && lb == 0 {
		return 1
	}
	maxLen := la
	if lb > maxLen {
		maxLen = lb
	}
	if maxLen == 0 {
		return 0
	}
	// 滚动数组 Levenshtein
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			m := prev[j] + 1
			if curr[j-1]+1 < m {
				m = curr[j-1] + 1
			}
			if prev[j-1]+cost < m {
				m = prev[j-1] + cost
			}
			curr[j] = m
		}
		prev, curr = curr, prev
	}
	return 1 - float64(prev[lb])/float64(maxLen)
}

// SheetAnswer 抽查用答卷答案行
type SheetAnswer struct {
	SheetId uint
	Answer  string
}

// SuspiciousPair 可疑相似对
type SuspiciousPair struct {
	SheetA uint    `json:"sheetA"`
	SheetB uint    `json:"sheetB"`
	Ratio  float64 `json:"ratio"`
}

// findSuspiciousPairs 相似对抽查（纯函数，测试锚点）：按答卷两两聚合（同一答卷主观题
// 相似度取最大值），≥阈值判可疑，相似度降序，上限 cheatMaxPairs 防组合爆炸
func findSuspiciousPairs(answers []SheetAnswer, sampleSize int) []SuspiciousPair {
	if sampleSize <= 0 {
		sampleSize = 10
	}
	if len(answers) < 2 {
		return []SuspiciousPair{}
	}
	// 答卷聚合：同答卷多题取最大相似度
	best := map[[2]uint]float64{}
	for i := 0; i < len(answers); i++ {
		for j := i + 1; j < len(answers); j++ {
			if answers[i].SheetId == answers[j].SheetId {
				continue
			}
			ratio := editSimilarity(answers[i].Answer, answers[j].Answer)
			if ratio < cheatSimThreshold {
				continue
			}
			key := [2]uint{answers[i].SheetId, answers[j].SheetId}
			if key[0] > key[1] {
				key[0], key[1] = key[1], key[0]
			}
			if ratio > best[key] {
				best[key] = ratio
			}
		}
	}
	out := make([]SuspiciousPair, 0, len(best))
	for key, ratio := range best {
		out = append(out, SuspiciousPair{SheetA: key[0], SheetB: key[1], Ratio: ratio})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Ratio > out[j].Ratio })
	if len(out) > sampleSize {
		out = out[:sampleSize]
	}
	return out
}

// ReportBehavior 行为上报（答卷存在性校验 + 白名单）
func (s *ExamService) ReportBehavior(sheetId uint, event, detail string) error {
	if !ValidBehaviorEvent(event) {
		return fmt.Errorf("[%d]不支持的行为事件: %s", ErrCodeCheatBad, event)
	}
	var sheet model.ExamAnswerSheet
	if err := global.GVA_DB.First(&sheet, sheetId).Error; err != nil {
		return fmt.Errorf("[%d]答卷不存在", ErrCodeSheetBad)
	}
	return global.GVA_DB.Create(&model.ExamBehavior{SheetId: sheetId, Event: event, Detail: detail}).Error
}

// GetBehaviorStats 行为计数（按事件聚合）
func (s *ExamService) GetBehaviorStats(sheetId uint) (map[string]int64, error) {
	type row struct {
		Event string
		Cnt   int64
	}
	var rows []row
	err := global.GVA_DB.Model(&model.ExamBehavior{}).
		Select("event, COUNT(*) as cnt").Where("sheet_id = ?", sheetId).
		Group("event").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := map[string]int64{}
	for _, r := range rows {
		out[r.Event] = r.Cnt
	}
	return out, nil
}

// CheckSimilarity 相似度抽查（同考次已交卷答卷主观题答案两两比对）
func (s *ExamService) CheckSimilarity(examId uint, sampleSize int) ([]SuspiciousPair, error) {
	var sheets []model.ExamAnswerSheet
	if err := global.GVA_DB.Where("exam_id = ? AND status = 'submitted'", examId).Find(&sheets).Error; err != nil {
		return nil, err
	}
	if len(sheets) < 2 {
		return []SuspiciousPair{}, nil
	}
	// 收集各卷主观题答案
	answers := make([]SheetAnswer, 0, len(sheets)*2)
	for _, sheet := range sheets {
		ansMap := map[string]string{}
		_ = parseSheetAnswers(sheet.AnswersJSON, ansMap)
		if len(ansMap) == 0 {
			continue
		}
		var qs []model.ExamQuestion
		var exam model.ExamExam
		if err := global.GVA_DB.First(&exam, sheet.ExamId).Error; err != nil {
			continue
		}
		var paper model.ExamPaper
		if err := global.GVA_DB.First(&paper, exam.PaperId).Error; err != nil {
			continue
		}
		if paper.QuestionIds == "" {
			continue
		}
		global.GVA_DB.Where("id IN ? AND type IN ?", paperQuestionIds(paper.QuestionIds), []string{"short", "practice"}).Find(&qs)
		for _, q := range qs {
			if a := ansMap[fmt.Sprintf("%d", q.ID)]; strings.TrimSpace(a) != "" {
				answers = append(answers, SheetAnswer{SheetId: sheet.ID, Answer: a})
			}
		}
	}
	return findSuspiciousPairs(answers, sampleSize), nil
}

// parseSheetAnswers 答卷答案 JSON 解析助手
func parseSheetAnswers(raw string, out map[string]string) error {
	return json.Unmarshal([]byte(raw), &out)
}

// BehaviorIn 行为上报请求
type BehaviorIn struct {
	SheetId uint   `json:"sheetId" binding:"required"`
	Event   string `json:"event" binding:"required"`
	Detail  string `json:"detail"`
}

// SimCheckIn 相似度抽查请求
type SimCheckIn struct {
	ExamId     uint `json:"examId" binding:"required"`
	SampleSize int  `json:"sampleSize"`
}
