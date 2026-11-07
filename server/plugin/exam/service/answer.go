// Package service exam：答题链路（券码登录→拉题脱敏→答案幂等保存→交卷防重）
// 口径：拉题剥除标准答案；保存走乐观锁版本号（幂等：同版本重复保存返回成功）；
// 保存/交卷均校验考试进行中 + 时间窗；交卷状态机 ongoing→submitted 行锁防重。
package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/exam/model"
)

// 错误码（exam 段续用）
const (
	ErrCodeSheetBad     = 2411
	ErrCodeSheetWindow  = 2412
	ErrCodeSheetVer     = 2413
)

// stripAnswers 拉题脱敏（纯函数，测试锚点）：剥除标准答案
func stripAnswers(qs []model.ExamQuestion) []model.ExamQuestion {
	out := make([]model.ExamQuestion, len(qs))
	for i, q := range qs {
		q.Answer = ""
		out[i] = q
	}
	return out
}

// checkVersion 乐观锁版本校验（纯函数：incoming<0 视为强制；不等报 2413）
func checkVersion(current, incoming int) error {
	if incoming < 0 {
		return nil
	}
	if incoming != current {
		return fmt.Errorf("[%d]版本冲突（当前 v%d，提交 v%d），请刷新后重试", ErrCodeSheetVer, current, incoming)
	}
	return nil
}

// withinExamWindow 作答时间窗校验（纯函数）
func withinExamWindow(start, end, now time.Time) error {
	if now.Before(start) {
		return fmt.Errorf("[%d]考试未开始", ErrCodeSheetWindow)
	}
	if now.After(end) {
		return fmt.Errorf("[%d]考试已结束", ErrCodeSheetWindow)
	}
	return nil
}

// LoginByTicket 券码登录：校验考生与考试 → 幂等创建答卷
func (s *ExamService) LoginByTicket(ticket, userKey string) (*model.ExamAnswerSheet, *model.ExamExam, error) {
	ticket = strings.ToUpper(strings.TrimSpace(ticket))
	if ticket == "" {
		return nil, nil, fmt.Errorf("[%d]券码不能为空", ErrCodeSheetBad)
	}
	var cand model.ExamCandidate
	if err := global.GVA_DB.Where("ticket_code = ?", ticket).First(&cand).Error; err != nil {
		return nil, nil, fmt.Errorf("[%d]券码无效", ErrCodeSheetBad)
	}
	var exam model.ExamExam
	if err := global.GVA_DB.First(&exam, cand.ExamId).Error; err != nil {
		return nil, nil, fmt.Errorf("[%d]考试不存在", ErrCodeExamNotFound)
	}
	if exam.Status != "ongoing" {
		return nil, nil, fmt.Errorf("[%d]考试未在进行中（当前 %s）", ErrCodeSheetWindow, exam.Status)
	}
	if err := withinExamWindow(exam.StartTime, exam.EndTime, time.Now()); err != nil {
		return nil, nil, err
	}
	// 关联 userKey（幂等）
	if userKey != "" && cand.UserKey != userKey {
		global.GVA_DB.Model(&cand).Update("user_key", userKey)
	}
	var sheet model.ExamAnswerSheet
	err := global.GVA_DB.Where("exam_id = ? AND candidate_id = ?", exam.ID, cand.ID).First(&sheet).Error
	if err == gorm.ErrRecordNotFound {
		sheet = model.ExamAnswerSheet{
			ExamId: exam.ID, CandidateId: cand.ID,
			AnswersJSON: "{}", StartedAt: time.Now(), Status: "ongoing",
		}
		if err := global.GVA_DB.Create(&sheet).Error; err != nil {
			return nil, nil, err
		}
	} else if err != nil {
		return nil, nil, err
	}
	global.GVA_DB.Model(&cand).Update("status", "entered")
	return &sheet, &exam, nil
}

// GetSheetQuestions 拉题（答卷 + 试卷题目脱敏 + 已存答案）
func (s *ExamService) GetSheetQuestions(sheetId uint) ([]model.ExamQuestion, *model.ExamAnswerSheet, map[string]string, error) {
	var sheetRow model.ExamAnswerSheet
	if err := global.GVA_DB.First(&sheetRow, sheetId).Error; err != nil {
		return nil, nil, nil, fmt.Errorf("[%d]答卷不存在", ErrCodeSheetBad)
	}
	var exam model.ExamExam
	if err := global.GVA_DB.First(&exam, sheetRow.ExamId).Error; err != nil {
		return nil, nil, nil, err
	}
	var paper model.ExamPaper
	if err := global.GVA_DB.First(&paper, exam.PaperId).Error; err != nil {
		return nil, nil, nil, err
	}
	if paper.QuestionIds == "" {
		return []model.ExamQuestion{}, &sheetRow, map[string]string{}, nil
	}
	idStrs := strings.Split(paper.QuestionIds, ",")
	ids := make([]uint, 0, len(idStrs))
	for _, idStr := range idStrs {
		if id, e := strconv.ParseUint(strings.TrimSpace(idStr), 10, 64); e == nil {
			ids = append(ids, uint(id))
		}
	}
	var questions []model.ExamQuestion
	if err := global.GVA_DB.Where("id IN ?", ids).Find(&questions).Error; err != nil {
		return nil, nil, nil, err
	}
	questions = stripAnswers(questions)
	answers := map[string]string{}
	_ = json.Unmarshal([]byte(sheetRow.AnswersJSON), &answers)
	return questions, &sheetRow, answers, nil
}

// SaveAnswers 答案保存（乐观锁幂等；进行中+时间窗校验）
func (s *ExamService) SaveAnswers(sheetId uint, answers map[string]string, version int) (*model.ExamAnswerSheet, error) {
	var sheet model.ExamAnswerSheet
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&sheet, sheetId).Error; err != nil {
			return fmt.Errorf("[%d]答卷不存在", ErrCodeSheetBad)
		}
		if sheet.Status != "ongoing" {
			return fmt.Errorf("[%d]答卷已交卷，不可再保存", ErrCodeSheetBad)
		}
		var exam model.ExamExam
		if err := tx.First(&exam, sheet.ExamId).Error; err != nil {
			return err
		}
		if err := withinExamWindow(exam.StartTime, exam.EndTime, time.Now()); err != nil {
			return err
		}
		if err := checkVersion(sheet.Version, version); err != nil {
			return err
		}
		b, _ := json.Marshal(answers)
		return tx.Model(&sheet).Updates(map[string]any{"answers_json": string(b), "version": sheet.Version + 1}).Error
	})
	if err != nil {
		return nil, err
	}
	sheet.Version++
	return &sheet, nil
}

// SubmitSheet 交卷（行锁状态机防重；幂等返回已交卷答卷）
func (s *ExamService) SubmitSheet(sheetId uint) (*model.ExamAnswerSheet, error) {
	var result *model.ExamAnswerSheet
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		var sheet model.ExamAnswerSheet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&sheet, sheetId).Error; err != nil {
			return fmt.Errorf("[%d]答卷不存在", ErrCodeSheetBad)
		}
		if sheet.Status == "submitted" {
			result = &sheet // 幂等
			return nil
		}
		now := time.Now()
		if err := tx.Model(&sheet).Updates(map[string]any{
			"status": "submitted", "submitted_at": now,
		}).Error; err != nil {
			return err
		}
		sheet.Status, sheet.SubmittedAt = "submitted", &now
		result = &sheet
		return nil
	})
	if err != nil {
		return nil, err
	}
	global.GVA_DB.Model(&model.ExamCandidate{}).Where("id = ?", result.CandidateId).Update("status", "submitted")
	return result, nil
}

// TicketLoginIn 券码登录请求
type TicketLoginIn struct {
	Ticket  string `json:"ticket" binding:"required"`
	UserKey string `json:"userKey"`
}

// SaveAnswersIn 答案保存请求（version=-1 强制）
type SaveAnswersIn struct {
	SheetId uint              `json:"sheetId" binding:"required"`
	Answers map[string]string `json:"answers"`
	Version int               `json:"version"`
}

// SubmitIn 交卷请求
type SubmitIn struct {
	SheetId uint `json:"sheetId" binding:"required"`
}
