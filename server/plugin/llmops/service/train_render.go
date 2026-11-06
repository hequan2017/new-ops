// Package service llmops：训练脚本渲染器（要求(4)）——超参白名单 → 命令行渲染 + dry-run
// 口径：超参仅接受白名单字段（未知字段忽略）；字符串插值一律经 sanitizePath 去 shell 元字符
// （; & | ` $ ( ) < > 换行）；数字由强类型格式化天然安全；本地 dry-run 仅渲染落 LogText 不执行。
package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const ErrCodeTrainRender = 2257

// TrainHyperParams 超参白名单结构
type TrainHyperParams struct {
	LearningRate         float64 `json:"learningRate"`
	Epochs               int     `json:"epochs"`
	BatchSize            int     `json:"batchSize"`
	GradientAccumSteps   int     `json:"gradientAccumSteps"`
	MaxSeqLen            int     `json:"maxSeqLen"`
	WarmupRatio          float64 `json:"warmupRatio"`
	LoraR                int     `json:"loraR"`
	LoraAlpha            int     `json:"loraAlpha"`
	LoraDropout          float64 `json:"loraDropout"`
	MixedPrecision       string  `json:"mixedPrecision"` // bf16/fp16/off
	GPUNum               int     `json:"gpuNum"`
	OutputDir            string  `json:"outputDir"`
}

// sanitizeTrainArg 字符串插值防注入（shell 元字符与控制符拒绝）
func sanitizeTrainArg(s string) error {
	if strings.ContainsAny(s, ";&|`$()<>\n\r\\\"'") {
		return fmt.Errorf("[%d]参数含非法字符: %q", ErrCodeTrainRender, s)
	}
	return nil
}

// ParseHyper 解析超参（缺省值补齐；未知字段忽略）
func ParseHyper(raw string) (*TrainHyperParams, error) {
	p := &TrainHyperParams{}
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), p); err != nil {
			return nil, fmt.Errorf("[%d]超参解析失败: %w", ErrCodeTrainRender, err)
		}
	}
	// 缺省值
	if p.LearningRate <= 0 {
		p.LearningRate = 2e-5
	}
	if p.Epochs <= 0 {
		p.Epochs = 3
	}
	if p.BatchSize <= 0 {
		p.BatchSize = 8
	}
	if p.GradientAccumSteps <= 0 {
		p.GradientAccumSteps = 1
	}
	if p.MaxSeqLen <= 0 {
		p.MaxSeqLen = 2048
	}
	if p.WarmupRatio < 0 {
		p.WarmupRatio = 0
	}
	if p.MixedPrecision == "" {
		p.MixedPrecision = "bf16"
	}
	if p.GPUNum <= 0 {
		p.GPUNum = 1
	}
	if p.OutputDir == "" {
		p.OutputDir = "/data/outputs"
	}
	if err := sanitizeTrainArg(p.OutputDir); err != nil {
		return nil, err
	}
	if err := sanitizeTrainArg(p.MixedPrecision); err != nil {
		return nil, err
	}
	if p.MixedPrecision != "bf16" && p.MixedPrecision != "fp16" && p.MixedPrecision != "off" {
		return nil, fmt.Errorf("[%d]mixedPrecision 仅支持 bf16/fp16/off", ErrCodeTrainRender)
	}
	if p.LoraR <= 0 {
		p.LoraR = 8
	}
	if p.LoraAlpha <= 0 {
		p.LoraAlpha = 16
	}
	if p.LoraDropout <= 0 || p.LoraDropout >= 1 { // 0 视为未设置（显式 0 与缺省不可区分）
		p.LoraDropout = 0.05
	}
	return p, nil
}

// RenderTrainCommand 渲染训练命令（sft 全参/LoRA 两套模板；其余类型由后续条目扩展）
func RenderTrainCommand(jobType string, p *TrainHyperParams, datasetPath, baseModel string) (string, error) {
	if jobType != "sft_full" && jobType != "sft_lora" {
		return "", fmt.Errorf("[%d]任务类型 %s 的渲染器尚未接入", ErrCodeTrainRender, jobType)
	}
	if err := sanitizeTrainArg(datasetPath); err != nil {
		return "", err
	}
	if err := sanitizeTrainArg(baseModel); err != nil {
		return "", err
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "torchrun --nproc_per_node=%d train_%s.py", p.GPUNum, jobType)
	fmt.Fprintf(&sb, " --model_name_or_path %s", baseModel)
	fmt.Fprintf(&sb, " --dataset_path %s", datasetPath)
	fmt.Fprintf(&sb, " --learning_rate %g --epochs %d --batch_size %d --grad_accum %d",
		p.LearningRate, p.Epochs, p.BatchSize, p.GradientAccumSteps)
	fmt.Fprintf(&sb, " --max_seq_len %d --warmup_ratio %g --precision %s --output_dir %s",
		p.MaxSeqLen, p.WarmupRatio, p.MixedPrecision, p.OutputDir)
	if jobType == "sft_lora" {
		fmt.Fprintf(&sb, " --lora_r %d --lora_alpha %d --lora_dropout %g", p.LoraR, p.LoraAlpha, p.LoraDropout)
	}
	return sb.String(), nil
}

// RunTrainDryRun 任务 dry-run：渲染命令落 LogText（状态不变，真实执行由启动条目接入）
func (s *LlmopsServiceStruct) RunTrainDryRun(jobId uint) (*model.LlmopsTrainJob, error) {
	var job model.LlmopsTrainJob
	if err := global.GVA_DB.First(&job, jobId).Error; err != nil {
		return nil, fmt.Errorf("[%d]任务不存在", ErrCodeTrainNotFound)
	}
	if job.Type != "sft_full" && job.Type != "sft_lora" {
		return nil, fmt.Errorf("[%d]任务类型 %s 的渲染器尚未接入", ErrCodeTrainRender, job.Type)
	}
	p, err := ParseHyper(job.HyperJSON)
	if err != nil {
		return nil, err
	}
	datasetPath := fmt.Sprintf("/data/datasets/%d", job.DatasetId)
	cmd, err := RenderTrainCommand(job.Type, p, datasetPath, job.BaseModel)
	if err != nil {
		return nil, err
	}
	if job.BaseModel == "" {
		return nil, fmt.Errorf("[%d]基座模型不能为空", ErrCodeTrainRender)
	}
	logText := "# dry-run 渲染（本地模式不执行）\n" + cmd
	if err := global.GVA_DB.Model(&job).Update("log_text", logText).Error; err != nil {
		return nil, err
	}
	job.LogText = logText
	return &job, nil
}
