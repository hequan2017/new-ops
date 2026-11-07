// Package service exam：断线续考（采购件④）——终端重进恢复答卷 + 单场时长截断
// 口径：resume 以券码为凭证（幂等复用 I05 登录链），一次调用返回题目+已存答案+剩余秒数；
// 剩余时长双口径取小：考试截止时间 与 单场时长（DurationMin 自答卷开始起算）。
package service

import (
	"fmt"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/exam/model"
)

// remainingSeconds 剩余作答秒数（纯函数，测试锚点）：
// min(考试截止-now, startedAt+duration-now)，下限 0；startedAt 零值仅按考试截止。
func remainingSeconds(examEnd time.Time, startedAt time.Time, durationMin int, now time.Time) int64 {
	remain := examEnd.Sub(now)
	if !startedAt.IsZero() && durationMin > 0 {
		personal := startedAt.Add(time.Duration(durationMin) * time.Minute).Sub(now)
		if personal < remain {
			remain = personal
		}
	}
	if remain < 0 {
		return 0
	}
	return int64(remain.Seconds())
}

// ResumeSession 断线续考：券码幂等登录 + 拉题恢复 + 剩余时长
func (s *ExamService) ResumeSession(ticket, userKey string) (questions []model.ExamQuestion, sheet *model.ExamAnswerSheet, answers map[string]string, remainSec int64, err error) {
	sheet, exam, err := s.LoginByTicket(ticket, userKey)
	if err != nil {
		return nil, nil, nil, 0, err
	}
	questions, sheet, answers, err = s.GetSheetQuestions(sheet.ID)
	if err != nil {
		return nil, nil, nil, 0, err
	}
	remainSec = remainingSeconds(exam.EndTime, sheet.StartedAt, exam.DurationMin, time.Now())
	if remainSec <= 0 {
		return nil, nil, nil, 0, fmt.Errorf("[%d]作答时长已用尽", ErrCodeSheetWindow)
	}
	return questions, sheet, answers, remainSec, nil
}

// ResumeIn 续考请求
type ResumeIn struct {
	Ticket  string `json:"ticket" binding:"required"`
	UserKey string `json:"userKey"`
}
