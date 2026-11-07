// Package service llmops：训练任务费用/时长估算（备用池 P07）
// 口径：经验模型估算——tokens ≈ 样本数 × 平均长度(字符/1.6 中文token经验值) × epochs；
// 时长(小时) ≈ tokens / 吞吐(默认 3000 token/s，GPU 数线性加速×0.85 并行效率)；
// 费用 ≈ 时长 × 卡数 × 单卡时价(可配，默认 6 元/时)。纯函数全链可测，供前端展示参考值。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 估算口径常量
const (
	estimateCharsPerToken = 1.6   // 中文经验：约 1.6 字符/token
	estimateThroughput    = 3000  // 单卡吞吐 token/s（经验默认）
	estimateParallelEff   = 0.85  // 多卡并行效率
	estimateUnitPrice     = 6.0   // 单卡时价（元/时，可配）
)

// TrainEstimate 估算结果
type TrainEstimate struct {
	Tokens      int64   `json:"tokens"`
	Hours       float64 `json:"hours"`
	Cost        float64 `json:"cost"`
	SampleCount int     `json:"sampleCount"`
}

// avgSampleChars 数据集平均样本长度（字符）
func avgSampleChars(content string) float64 {
	lines := strings.Split(content, "\n")
	total := 0
	n := 0
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		total += len([]rune(l))
		n++
	}
	if n == 0 {
		return 0
	}
	return float64(total) / float64(n)
}

// EstimateTrain 训练估算（纯函数，测试锚点）
func EstimateTrain(sampleCount int, avgChars float64, epochs, gpuNum int, unitPrice float64) (*TrainEstimate, error) {
	if sampleCount <= 0 {
		return nil, fmt.Errorf("[%d]样本数为 0，无法估算", ErrCodeDatasetBad)
	}
	if epochs <= 0 {
		epochs = 3
	}
	if gpuNum <= 0 {
		gpuNum = 1
	}
	if unitPrice <= 0 {
		unitPrice = estimateUnitPrice
	}
	tokens := int64(float64(sampleCount) * avgChars / estimateCharsPerToken * float64(epochs))
	throughput := estimateThroughput * float64(gpuNum) * estimateParallelEff
	hours := float64(tokens) / throughput / 3600
	return &TrainEstimate{
		Tokens:      tokens,
		Hours:       hours,
		Cost:        hours * float64(gpuNum) * unitPrice,
		SampleCount: sampleCount,
	}, nil
}

// GetTrainEstimate 任务估算入口：datasetId 引用的数据集实际内容统计
func (s *LlmopsServiceStruct) GetTrainEstimate(jobId uint) (*TrainEstimate, error) {
	var job model.LlmopsTrainJob
	if err := global.GVA_DB.First(&job, jobId).Error; err != nil {
		return nil, fmt.Errorf("[%d]任务不存在", ErrCodeTrainNotFound)
	}
	if job.DatasetId == 0 {
		return nil, fmt.Errorf("[%d]任务未挂数据集，无法估算", ErrCodeDatasetBad)
	}
	var d model.LlmopsDataset
	if err := global.GVA_DB.First(&d, job.DatasetId).Error; err != nil {
		return nil, fmt.Errorf("[%d]数据集不存在", ErrCodeDatasetBad)
	}
	if strings.TrimSpace(d.Content) == "" {
		return nil, fmt.Errorf("[%d]数据集内容为空", ErrCodeDatasetBad)
	}
	hyper, _ := ParseHyper(job.HyperJSON)
	sampleCount := CountDatasetSamples(d.Content, "jsonl")
	return EstimateTrain(sampleCount, avgSampleChars(d.Content), hyper.Epochs, hyper.GPUNum, estimateUnitPrice)
}
