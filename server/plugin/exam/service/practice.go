// Package service exam：实操题执行（采购件④）——算子/智能体双通道 + 过程留痕
// 口径：operator 通道跨插件调用 oplib 执行引擎（schema 校验+内置/脚本全量）；
// agent 通道跨插件调用 llmops 网关单轮作答（多轮工具回路走调试面，注明口径）；
// 执行输出自动写回答卷（questionId → 输出前 1000 rune，version=-1 强制保存）作为判分依据。
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/exam/model"
	oplibService "github.com/hequan2017/new-ops/server/plugin/oplib/service"
	llmopsService "github.com/hequan2017/new-ops/server/plugin/llmops/service"
)

// 错误码（exam 段续用）
const (
	ErrCodePracticeBad = 2420
	practiceOutputMax  = 1000
)

// PracticeChannels 实操通道白名单
var PracticeChannels = []string{"operator", "agent"}

// ValidPracticeChannel 通道白名单
func ValidPracticeChannel(c string) bool {
	return c == "operator" || c == "agent"
}

// truncatePracticeOutput 输出截断（纯函数：写回答卷的摘要口径）
func truncatePracticeOutput(s string) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= practiceOutputMax {
		return string(r)
	}
	return string(r[:practiceOutputMax]) + "…(截断)"
}

// RunPractice 实操执行：校验→执行→留痕→写回答卷
func (s *ExamService) RunPractice(sheetId, questionId uint, channel, target, input string) (*model.ExamPracticeLog, error) {
	if !ValidPracticeChannel(channel) {
		return nil, fmt.Errorf("[%d]通道须为 operator/agent", ErrCodePracticeBad)
	}
	if strings.TrimSpace(target) == "" {
		return nil, fmt.Errorf("[%d]执行目标不能为空", ErrCodePracticeBad)
	}
	var sheet model.ExamAnswerSheet
	if err := global.GVA_DB.First(&sheet, sheetId).Error; err != nil {
		return nil, fmt.Errorf("[%d]答卷不存在", ErrCodeSheetBad)
	}
	if sheet.Status != "ongoing" {
		return nil, fmt.Errorf("[%d]答卷已交卷", ErrCodeSheetBad)
	}
	var exam model.ExamExam
	if err := global.GVA_DB.First(&exam, sheet.ExamId).Error; err != nil {
		return nil, err
	}
	if err := withinExamWindow(exam.StartTime, exam.EndTime, time.Now()); err != nil {
		return nil, err
	}

	log := &model.ExamPracticeLog{
		SheetId: sheetId, QuestionId: questionId, Channel: channel, Target: target, Input: input,
	}
	started := time.Now()
	switch channel {
	case "operator":
		runLog, err := oplibService.OplibApp.ExecuteOperator(parseUintTarget(target), input)
		log.LatencyMs = time.Since(started).Milliseconds()
		if err != nil {
			log.Ok = false
			log.ErrMsg = truncateSample(err.Error())
			if runLog != nil && runLog.ResultJSON != "" {
				log.Output = runLog.ResultJSON
			}
		} else {
			log.Ok = true
			log.Output = runLog.ResultJSON
		}
	default: // agent：单轮对话作答（模型由能力调度选出）
		ctx, cancel := contextWithTimeout()
		defer cancel()
		m, errSel := llmopsService.LlmopsService.SelectModel("chat")
		if errSel != nil {
			log.Ok, log.ErrMsg, log.LatencyMs = false, truncateSample(errSel.Error()), time.Since(started).Milliseconds()
			break
		}
		content, _, errChat := llmopsService.LlmopsService.ChatCompletion(ctx, m.ID,
			[]llmopsService.ChatMessage{{Role: "user", Content: input}}, nil)
		log.LatencyMs = time.Since(started).Milliseconds()
		if errChat != nil {
			log.Ok, log.ErrMsg = false, truncateSample(errChat.Error())
		} else {
			log.Ok, log.Output = true, content
		}
	}
	if err := global.GVA_DB.Create(log).Error; err != nil {
		return nil, err
	}
	// 输出写回答卷作为判分依据（强制版本）
	if log.Ok && log.Output != "" {
		answers := map[string]string{}
		_, _, saved, err := s.GetSheetQuestions(sheetId)
		if err == nil {
			for k, v := range saved {
				answers[k] = v
			}
		}
		answers[fmt.Sprintf("%d", questionId)] = truncatePracticeOutput(log.Output)
		if _, err := s.SaveAnswers(sheetId, answers, -1); err != nil {
			global.GVA_LOG.Error(fmt.Sprintf("llmops exam 实操输出写回失败: %v", err))
		}
	}
	return log, nil
}

// PracticeIn 实操执行请求
type PracticeIn struct {
	SheetId    uint   `json:"sheetId" binding:"required"`
	QuestionId uint   `json:"questionId" binding:"required"`
	Channel    string `json:"channel" binding:"required"`
	Target     string `json:"target" binding:"required"`
	Input      string `json:"input"`
}

// GetPracticeLogs 实操过程日志（终端面：sheetId 即凭证口径，仅返回本卷日志）
func (s *ExamService) GetPracticeLogs(sheetId, questionId uint) ([]model.ExamPracticeLog, error) {
	db := global.GVA_DB.Where("sheet_id = ?", sheetId)
	if questionId > 0 {
		db = db.Where("question_id = ?", questionId)
	}
	var list []model.ExamPracticeLog
	err := db.Order("id desc").Limit(20).Find(&list).Error
	return list, err
}
