// Package service exam：考试场次管理（四态状态机：草稿→发布→进行中→已结束）
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/exam/model"
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
		return fmt.Errorf("[%d]开考时间必须早于截止时间", ErrCodeExamWindow)
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
	db := global.GVA_DB.Model(&model.ExamExam{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	var list []model.ExamExam
	err := db.Order("start_time desc").Find(&list).Error
	return list, err
}

// SaveExam 新建/更新考试（试卷须 ready；时间窗校验）
func (s *ExamService) SaveExam(in *ExamIn) (*model.ExamExam, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("[%d]考试名必填", ErrCodeExamWindow)
	}
	if err := ValidateExamWindow(in.StartTime, in.EndTime); err != nil {
		return nil, err
	}
	if in.DurationMin <= 0 {
		in.DurationMin = 60
	}
	var paper model.ExamPaper
	if err := global.GVA_DB.First(&paper, in.PaperId).Error; err != nil {
		return nil, fmt.Errorf("[%d]试卷不存在", ErrCodePaperNotFound)
	}
	if paper.Status != "ready" {
		return nil, fmt.Errorf("[%d]试卷未组卷（状态 %s），请先组卷", ErrCodeExamWindow, paper.Status)
	}
	var count int64
	q := global.GVA_DB.Model(&model.ExamExam{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]考试名已存在", ErrCodeExamWindow)
	}
	if in.ID > 0 {
		var e model.ExamExam
		if err := global.GVA_DB.First(&e, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]考试不存在", ErrCodeExamNotFound)
		}
		if e.Status != "draft" {
			return nil, fmt.Errorf("[%d]仅草稿态可编辑（当前 %s）", ErrCodeExamTrans, e.Status)
		}
		err := global.GVA_DB.Model(&e).Updates(map[string]any{
			"name": in.Name, "paper_id": in.PaperId,
			"start_time": in.StartTime, "end_time": in.EndTime, "duration_min": in.DurationMin,
		}).Error
		if err != nil {
			return nil, err
		}
		return &e, nil
	}
	e := &model.ExamExam{
		Name: in.Name, PaperId: in.PaperId, StartTime: in.StartTime,
		EndTime: in.EndTime, DurationMin: in.DurationMin, Status: "draft",
	}
	if err := global.GVA_DB.Create(e).Error; err != nil {
		return nil, err
	}
	return e, nil
}

// TransitionExam 状态流转（published 须试卷 ready；ongoing 不早于开考时间软校验）
func (s *ExamService) TransitionExam(id uint, to string) (*model.ExamExam, error) {
	var e model.ExamExam
	if err := global.GVA_DB.First(&e, id).Error; err != nil {
		return nil, fmt.Errorf("[%d]考试不存在", ErrCodeExamNotFound)
	}
	if !CanTransitionExam(e.Status, to) {
		return nil, fmt.Errorf("[%d]非法流转 %s → %s", ErrCodeExamTrans, e.Status, to)
	}
	if to == "ongoing" && time.Now().Before(e.StartTime) {
		return nil, fmt.Errorf("[%d]未到开考时间（%s 开考）", ErrCodeExamWindow, e.StartTime.Format("2006-01-02 15:04"))
	}
	if err := global.GVA_DB.Model(&e).Update("status", to).Error; err != nil {
		return nil, err
	}
	e.Status = to
	return &e, nil
}

// DeleteExam 删除考试（仅草稿/已结束可删）
func (s *ExamService) DeleteExam(id uint) error {
	var e model.ExamExam
	if err := global.GVA_DB.First(&e, id).Error; err != nil {
		return fmt.Errorf("[%d]考试不存在", ErrCodeExamNotFound)
	}
	if e.Status == "ongoing" || e.Status == "published" {
		return fmt.Errorf("[%d]发布/进行中考试不可删除，请先结束", ErrCodeExamTrans)
	}
	return global.GVA_DB.Delete(&e).Error
}
