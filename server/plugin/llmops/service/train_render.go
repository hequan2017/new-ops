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
	ZeroStage            int     `json:"zeroStage"`              // DeepSpeed ZeRO 0-3（0 不启用）
	GradientCheckpointing bool   `json:"gradientCheckpointing"` // 梯度检查点换显存
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
	if p.ZeroStage < 0 || p.ZeroStage > 3 {
		return nil, fmt.Errorf("[%d]zeroStage 仅支持 0-3", ErrCodeTrainRender)
	}
	return p, nil
}

// RenderTrainCommand 渲染训练命令（sft 全参/LoRA 兼容入口，无奖励模型路径）
func RenderTrainCommand(jobType string, p *TrainHyperParams, datasetPath, baseModel string) (string, error) {
	return RenderTrainCommandEx(jobType, p, datasetPath, baseModel, "")
}

// RenderTrainCommandEx 渲染训练命令（四类已接入：sft_full/sft_lora/reward/rl；rl 附奖励模型路径）
func RenderTrainCommandEx(jobType string, p *TrainHyperParams, datasetPath, baseModel, rewardModelPath string) (string, error) {
	switch jobType {
	case "sft_full", "sft_lora", "reward", "rl":
	default:
		return "", fmt.Errorf("[%d]任务类型 %s 的渲染器尚未接入", ErrCodeTrainRender, jobType)
	}
	if err := sanitizeTrainArg(datasetPath); err != nil {
		return "", err
	}
	if err := sanitizeTrainArg(baseModel); err != nil {
		return "", err
	}
	if err := sanitizeTrainArg(rewardModelPath); err != nil {
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
	if p.ZeroStage > 0 { // 训练加速：DeepSpeed ZeRO 分片
		fmt.Fprintf(&sb, " --deepspeed --zero_stage %d", p.ZeroStage)
	}
	if p.GradientCheckpointing { // 训练加速：梯度检查点
		sb.WriteString(" --gradient_checkpointing")
	}
	switch jobType {
	case "sft_lora":
		fmt.Fprintf(&sb, " --lora_r %d --lora_alpha %d --lora_dropout %g", p.LoraR, p.LoraAlpha, p.LoraDropout)
	case "reward":
		fmt.Fprintf(&sb, " --loss sequence_pairwise") // 奖励模型配对损失口径
	case "rl":
		if rewardModelPath == "" {
			return "", fmt.Errorf("[%d]RL 阶段必须提供奖励模型路径", ErrCodeTrainRender)
		}
		fmt.Fprintf(&sb, " --reward_model_path %s", rewardModelPath)
	}
	return sb.String(), nil
}

// RunTrainDryRun 任务 dry-run：渲染命令落 LogText（状态不变，真实执行由启动条目接入）
func (s *LlmopsServiceStruct) RunTrainDryRun(jobId uint) (*model.LlmopsTrainJob, error) {
	var job model.LlmopsTrainJob
	if err := global.GVA_DB.First(&job, jobId).Error; err != nil {
		return nil, fmt.Errorf("[%d]任务不存在", ErrCodeTrainNotFound)
	}
	p, err := ParseHyper(job.HyperJSON)
	if err != nil {
		return nil, err
	}
	if job.BaseModel == "" {
		return nil, fmt.Errorf("[%d]基座模型不能为空", ErrCodeTrainRender)
	}
	// RLHF 两阶段链路校验与奖励模型路径解析
	rewardPath := ""
	if job.Type == "rl" {
		preJob, err := s.ResolveRlhfPreJob(jobId, job.PreJobId)
		if err != nil {
			return nil, err
		}
		preHyper, _ := ParseHyper(preJob.HyperJSON)
		rewardPath = preHyper.OutputDir // 奖励模型产物 = 前置 reward 任务输出目录
	}
	datasetPath := fmt.Sprintf("/data/datasets/%d", job.DatasetId)
	cmd, err := RenderTrainCommandEx(job.Type, p, datasetPath, job.BaseModel, rewardPath)
	if err != nil {
		return nil, err
	}
	logText := "# dry-run 渲染（本地模式不执行）\n" + cmd
	if err := global.GVA_DB.Model(&job).Update("log_text", logText).Error; err != nil {
		return nil, err
	}
	job.LogText = logText
	return &job, nil
}
