// Package service llmops：训练任务启动/停止/日志（要求(4) 任务管理）
// 口径：本地 dry-run 执行器——Start=渲染命令+状态机 queued→running，LogText 记录启动命令，
// 真实训练后端待接入（标注）；Stop=running→stopped 带原因；日志追加时间戳前缀。
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// appendTrainLog 追加带时间戳的日志段
func appendTrainLog(existing, section, body string) string {
	stamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("\n[%s] %s\n%s", stamp, section, body)
	if strings.TrimSpace(existing) == "" {
		return strings.TrimPrefix(entry, "\n")
	}
	return existing + entry
}

// StartTrainJob 启动任务：状态机推进 draft/failed/stopped→queued→running，渲染命令落日志
func (s *LlmopsServiceStruct) StartTrainJob(jobId uint) (*model.LlmopsTrainJob, error) {
	var job model.LlmopsTrainJob
	if err := global.GVA_DB.First(&job, jobId).Error; err != nil {
		return nil, fmt.Errorf("[%d]任务不存在", ErrCodeTrainNotFound)
	}
	switch job.Status {
	case "succeeded":
		return nil, fmt.Errorf("[%d]已成功任务不可再启动", ErrCodeTrainTrans)
	case "running":
		return nil, fmt.Errorf("[%d]任务已在运行中", ErrCodeTrainTrans)
	}
	// 渲染（RLHF 链路校验在渲染链内）
	rendered, err := s.RunTrainDryRun(jobId)
	if err != nil {
		return nil, err
	}
	// 状态推进：非 queued 先流转到 queued
	if job.Status != "queued" {
		if _, err := s.TransitionTrainJob(jobId, "queued", "启动入队"); err != nil {
			return nil, err
		}
	}
	if _, err := s.TransitionTrainJob(jobId, "running", ""); err != nil {
		return nil, err
	}
	startLog := appendTrainLog(rendered.LogText, "START（本地 dry-run 执行器，真实训练后端待接入）",
		"任务已置 running；训练命令见上方渲染结果。")
	if err := global.GVA_DB.Model(&job).Update("log_text", startLog).Error; err != nil {
		return nil, err
	}
	job.Status = "running"
	job.LogText = startLog
	return &job, nil
}

// StopTrainJob 停止任务：running→stopped 带原因
func (s *LlmopsServiceStruct) StopTrainJob(jobId uint, reason string) (*model.LlmopsTrainJob, error) {
	var job model.LlmopsTrainJob
	if err := global.GVA_DB.First(&job, jobId).Error; err != nil {
		return nil, fmt.Errorf("[%d]任务不存在", ErrCodeTrainNotFound)
	}
	if job.Status != "running" {
		return nil, fmt.Errorf("[%d]仅运行中任务可停止（当前 %s）", ErrCodeTrainTrans, job.Status)
	}
	updated, err := s.TransitionTrainJob(jobId, "stopped", reason)
	if err != nil {
		return nil, err
	}
	stopLog := appendTrainLog(updated.LogText, "STOP", "任务已停止："+reason)
	if err := global.GVA_DB.Model(&job).Update("log_text", stopLog).Error; err != nil {
		return nil, err
	}
	updated.LogText = stopLog
	return updated, nil
}

// GetTrainJobLog 任务执行日志
func (s *LlmopsServiceStruct) GetTrainJobLog(jobId uint) (string, string, error) {
	var job model.LlmopsTrainJob
	if err := global.GVA_DB.First(&job, jobId).Error; err != nil {
		return "", "", fmt.Errorf("[%d]任务不存在", ErrCodeTrainNotFound)
	}
	return job.LogText, job.Status, nil
}
