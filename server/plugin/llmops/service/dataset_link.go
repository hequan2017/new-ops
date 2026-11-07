// Package service llmops：数据集联动（采购件⑤）——任务引用校验 + 引用计数
// 口径：任务类型→允许的数据集用途映射（训练 sft/奖励与强化 rlhf/评估 eval/压缩无数据集）；
// dry-run 渲染前即校验存在性与用途匹配；引用计数供数据集页展示被哪些任务使用。
package service

import (
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// purposeForJob 任务类型→允许的数据集用途（纯函数，测试锚点）
func purposeForJob(jobType string) []string {
	switch jobType {
	case "sft_full", "sft_lora":
		return []string{"sft"}
	case "reward", "rl":
		return []string{"rlhf"}
	case "eval":
		return []string{"eval"}
	default: // compress 等无数据集依赖
		return nil
	}
}

// purposeAllowed 用途匹配（纯函数：allowed 为空表示不约束）
func purposeAllowed(actual string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, a := range allowed {
		if a == actual {
			return true
		}
	}
	return false
}

// ValidateDatasetForUse 数据集用途校验（存在性 + 匹配，供训练/评估任务调用）
func (s *LlmopsServiceStruct) ValidateDatasetForUse(datasetId uint, allowed []string) (*model.LlmopsDataset, error) {
	if datasetId == 0 {
		return nil, fmt.Errorf("[%d]未选择数据集", ErrCodeDatasetBad)
	}
	var d model.LlmopsDataset
	if err := global.GVA_DB.First(&d, datasetId).Error; err != nil {
		return nil, fmt.Errorf("[%d]数据集不存在", ErrCodeDatasetBad)
	}
	if !purposeAllowed(d.Purpose, allowed) {
		return nil, fmt.Errorf("[%d]数据集用途不匹配：任务需要 %v，实际 %s", ErrCodeDatasetBad, allowed, d.Purpose)
	}
	if strings_NoContent(d.Content) {
		return nil, fmt.Errorf("[%d]数据集内容为空", ErrCodeDatasetBad)
	}
	return &d, nil
}

func strings_NoContent(s string) bool { return len(s) == 0 }

// DatasetRefs 引用计数
type DatasetRefs struct {
	TrainJobs int64 `json:"trainJobs"`
	EvalRuns  int64 `json:"evalRuns"`
}

// GetDatasetRefs 数据集被引用计数（训练任务 + 评估任务）
func (s *LlmopsServiceStruct) GetDatasetRefs(datasetId uint) (*DatasetRefs, error) {
	var refs DatasetRefs
	if err := global.GVA_DB.Model(&model.LlmopsTrainJob{}).Where("dataset_id = ?", datasetId).Count(&refs.TrainJobs).Error; err != nil {
		return nil, err
	}
	if err := global.GVA_DB.Model(&model.LlmopsEvalRun{}).Where("dataset_id = ?", datasetId).Count(&refs.EvalRuns).Error; err != nil {
		return nil, err
	}
	return &refs, nil
}
