// Package service exam：考试终端接入（采购件④）——终端注册/心跳/可考列表
// 口径：终端面为独立公开路由（不走管理端 JWT），机器码白名单鉴权——请求头
// X-Terminal-Key 必须已注册且 online；可考列表仅 ongoing + 时间窗内场次。
package service

import (
	"fmt"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/exam/model"
)

// 错误码（exam 段续用）
const ErrCodeTermAuth = 2419

// VerifyTerminalKey 终端机器码白名单校验（存在 + online）
func (s *ExamService) VerifyTerminalKey(nodeKey string) (*model.ExamNode, error) {
	if nodeKey == "" {
		return nil, fmt.Errorf("[%d]终端标识缺失", ErrCodeTermAuth)
	}
	var node model.ExamNode
	if err := global.GVA_DB.Where("node_key = ?", nodeKey).First(&node).Error; err != nil {
		return nil, fmt.Errorf("[%d]终端未注册", ErrCodeTermAuth)
	}
	if node.Status != "online" {
		return nil, fmt.Errorf("[%d]终端已离线，请重新注册", ErrCodeTermAuth)
	}
	return &node, nil
}

// TerminalExam 终端视角的可考场次（脱敏：不暴露试卷 ID 与规则）
type TerminalExam struct {
	ExamId      uint   `json:"examId"`
	Name        string `json:"name"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	DurationMin int    `json:"durationMin"`
}

// filterOngoingExams 可考场次过滤（纯函数，测试锚点）：ongoing 且时间窗内
func filterOngoingExams(exams []model.ExamExam, now time.Time) []TerminalExam {
	out := make([]TerminalExam, 0, len(exams))
	for _, e := range exams {
		if e.Status != "ongoing" {
			continue
		}
		if now.Before(e.StartTime) || now.After(e.EndTime) {
			continue
		}
		out = append(out, TerminalExam{
			ExamId: e.ID, Name: e.Name, StartTime: e.StartTime,
			EndTime: e.EndTime, DurationMin: e.DurationMin,
		})
	}
	return out
}

// ListTerminalExams 终端拉取可考列表
func (s *ExamService) ListTerminalExams() ([]TerminalExam, error) {
	var exams []model.ExamExam
	if err := global.GVA_DB.Where("status = 'ongoing'").Find(&exams).Error; err != nil {
		return nil, err
	}
	return filterOngoingExams(exams, time.Now()), nil
}
