// Package service exam：考试场次管理（四态状态机：草稿→发布→进行中→已结束）
package service

import (
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/exam/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（exam 段续用）
const (
	ErrCodeExamNotFound = 2407
	ErrCodeExamTrans    = 2408
	ErrCodeExamWindow   = 2409
)

// examTransitions 状态机合法流转
var examTransitions = map[string][]string{
	"draft":     {"published", "finished"}, // finished 兜底废弃
	"published": {"ongoing", "draft", "finished"},
	"ongoing":   {"finished"},
	"finished":  {},
}

// CanTransitionExam 状态机流转校验（纯函数）
func CanTransitionExam(from, to string) bool {
	for _, t := range examTransitions[from] {
		if t == to {
			return true
		}
	}
	return false
}

// ValidateExamWindow 时间窗校验（纯函数：start<end）
func ValidateExamWindow(start, end time.Time) error {
	if !start.Before(end) {
		return examErr(ErrCodeExamWindow, "开考时间必须早于截止时间")
	}
	return nil
}

// ExamIn 考试保存入参
type ExamIn struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	PaperId     uint      `json:"paperId"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	DurationMin int       `json:"durationMin"`
}

// GetExams 考试列表（状态过滤）
func (s *ExamService) GetExams(status string) ([]model.ExamExam, error) {
	list, err := s.repo.ExamList(s.dbh(), status)
	if err != nil {
		return nil, errx.Wrap(ErrCodeExamStore, "查询考试列表失败", err)
	}
	return list, nil
}

// SaveExam 新建/更新考试（试卷须 ready；时间窗校验）
func (s *ExamService) SaveExam(in *ExamIn) (*model.ExamExam, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" {
		return nil, examErr(ErrCodeExamWindow, "考试名必填")
	}
	if err := ValidateExamWindow(in.StartTime, in.EndTime); err != nil {
		return nil, err
	}
	if in.DurationMin <= 0 {
		in.DurationMin = 60
	}
	db := s.dbh()
	paper, err := s.repo.FindPaper(db, in.PaperId)
	if err != nil {
		return nil, examErr(ErrCodePaperNotFound, "试卷不存在")
	}
	if paper.Status != "ready" {
		return nil, examErr(ErrCodeExamWindow, "试卷未组卷（状态 %s），请先组卷", paper.Status)
	}
	count, err := s.repo.CountExamsByName(db, in.Name, in.ID)
	if err != nil {
		return nil, errx.Wrap(ErrCodeExamStore, "查询考试重名失败", err)
	}
	if count > 0 {
		return nil, examErr(ErrCodeExamWindow, "考试名已存在")
	}
	if in.ID > 0 {
		e, err := s.repo.FindExam(db, in.ID)
		if err != nil {
			return nil, examErr(ErrCodeExamNotFound, "考试不存在")
		}
		if e.Status != "draft" {
			return nil, examErr(ErrCodeExamTrans, "仅草稿态可编辑（当前 %s）", e.Status)
		}
		if err := s.repo.UpdateExamColumns(db, e.ID, map[string]any{
			"name": in.Name, "paper_id": in.PaperId,
			"start_time": in.StartTime, "end_time": in.EndTime, "duration_min": in.DurationMin,
		}); err != nil {
			return nil, errx.Wrap(ErrCodeExamStore, "更新考试失败", err)
		}
		return e, nil
	}
	e := &model.ExamExam{
		Name: in.Name, PaperId: in.PaperId, StartTime: in.StartTime,
		EndTime: in.EndTime, DurationMin: in.DurationMin, Status: "draft",
	}
	if err := s.repo.CreateExam(db, e); err != nil {
		return nil, errx.Wrap(ErrCodeExamStore, "落库考试失败", err)
	}
	return e, nil
}

// TransitionExam 状态流转（published 须试卷 ready；ongoing 不早于开考时间软校验）
func (s *ExamService) TransitionExam(id uint, to string) (*model.ExamExam, error) {
	db := s.dbh()
	e, err := s.repo.FindExam(db, id)
	if err != nil {
		return nil, examErr(ErrCodeExamNotFound, "考试不存在")
	}
	if !CanTransitionExam(e.Status, to) {
		return nil, examErr(ErrCodeExamTrans, "非法流转 %s → %s", e.Status, to)
	}
	if to == "ongoing" && time.Now().Before(e.StartTime) {
		return nil, examErr(ErrCodeExamWindow, "未到开考时间（%s 开考）", e.StartTime.Format("2006-01-02 15:04"))
	}
	if err := s.repo.UpdateExamColumns(db, e.ID, map[string]any{"status": to}); err != nil {
		return nil, errx.Wrap(ErrCodeExamStore, "更新考试失败", err)
	}
	e.Status = to
	return e, nil
}

// DeleteExam 删除考试（仅草稿/已结束可删）
func (s *ExamService) DeleteExam(id uint) error {
	db := s.dbh()
	e, err := s.repo.FindExam(db, id)
	if err != nil {
		return examErr(ErrCodeExamNotFound, "考试不存在")
	}
	if e.Status == "ongoing" || e.Status == "published" {
		return examErr(ErrCodeExamTrans, "发布/进行中考试不可删除，请先结束")
	}
	if _, err := s.repo.DeleteExam(db, e.ID); err != nil {
		return errx.Wrap(ErrCodeExamStore, "删除考试失败", err)
	}
	return nil
}
