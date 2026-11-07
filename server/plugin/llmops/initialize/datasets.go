// Package initialize llmops：教育垂类种子数据集（采购件⑤ 内置示例）
// 口径：go:embed 三份 JSONL（sft/eval/exam 各 20 样本），FirstOrCreate 幂等导入，
// 样本数预计算；种子仅首次建，后续以页面编辑为准不被覆盖。
package initialize

import (
	"context"
	_ "embed"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

//go:embed resource/datasets/education-sft-demo.jsonl
var seedSFT string

//go:embed resource/datasets/education-eval-demo.jsonl
var seedEval string

//go:embed resource/datasets/education-exam-demo.jsonl
var seedExam string

// SeedDatasets 内置数据集种子（幂等）
func SeedDatasets(ctx context.Context) {
	presets := []model.LlmopsDataset{
		{
			Name: "教育垂类 SFT 示例", Purpose: "sft", Format: "jsonl", Version: "v1",
			License: "Apache-2.0", Content: seedSFT,
			Desc: "教育垂类大模型教学场景指令-回答对（内置示例）",
		},
		{
			Name: "教育垂类评测集示例", Purpose: "eval", Format: "jsonl", Version: "v1",
			License: "Apache-2.0", Content: seedEval,
			Desc: "教育垂类知识问答评测集（内置示例）",
		},
		{
			Name: "教育垂类考试题集示例", Purpose: "exam", Format: "jsonl", Version: "v1",
			License: "Apache-2.0", Content: seedExam,
			Desc: "五题型考题集（内置示例，可直接导入考试题库）",
		},
	}
	for i := range presets {
		presets[i].SampleCount = countSeedSamples(presets[i].Content)
		var exist model.LlmopsDataset
		err := global.GVA_DB.WithContext(ctx).Where("name = ?", presets[i].Name).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&presets[i]).Error; err != nil {
				global.GVA_LOG.Error("llmops 数据集种子失败: " + err.Error())
			}
		}
	}
}
