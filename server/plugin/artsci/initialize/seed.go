// Package initialize artsci：内置工具种子（幂等，M03 起逐条追加）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/artsci/model"
)

// SeedTools 内置工具种子（按名幂等：存在即跳过）
func SeedTools(ctx context.Context) {
	presets := []model.ArtsciTool{
		{
			Name: "Allen-Cahn 相场演化", Subject: "数学", Module: "allen-cahn", Status: "enabled",
			ParamsSchema: `{"epsilon":{"type":"number"},"gridN":{"type":"number"},"dt":{"type":"number"},"steps":{"type":"number"},"snapshotEvery":{"type":"number"}}`,
			Desc:         "一维 Allen-Cahn 方程显式有限差分求解（周期边界），返回演化帧序列与能量衰减（3.3-1）",
		},
		{
			Name: "DeepHPMs 层次预测", Subject: "数学", Module: "deep-hpms", Status: "enabled",
			ParamsSchema: `{"xs":{"type":"array"},"ys":{"type":"array"},"degree":{"type":"number"},"residualDegree":{"type":"number"},"predictXs":{"type":"array"}}`,
			Desc:         "两级层次预测：全局多项式最小二乘+残差局部修正，返回系数/三级 MSE 对比（3.3-4）",
		},
		{
			Name: "DeepONet 算子逼近", Subject: "数学", Module: "deep-onet", Status: "enabled",
			ParamsSchema: `{"trainFunctions":{"type":"array"},"trainYs":{"type":"array"},"trainTargets":{"type":"array"},"basisCount":{"type":"number"},"predictFunction":{"type":"array"},"predictYs":{"type":"array"}}`,
			Desc:         "勒让德 branch/trunk 内积结构算子学习，最小二乘拟合组合系数（3.3-6）",
		},
	}
	for i := range presets {
		var exist model.ArtsciTool
		err := global.GVA_DB.WithContext(ctx).Where("name = ?", presets[i].Name).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&presets[i]).Error; err != nil {
				global.GVA_LOG.Error("artsci 工具种子失败: " + err.Error())
			}
		}
	}
	seedSceneTemplates(ctx)
}

// seedSceneTemplates 文生图教学场景模板种子（幂等，需求 3.3-10/11/12/14/15）
func seedSceneTemplates(ctx context.Context) {
	presets := []model.ArtsciSceneTemplate{
		{
			Scene: "历史场景再现", Status: "enabled",
			PromptTemplate: "根据历史文献描述生成场景插画：{{text}}。要求：符合该历史时期的服饰、建筑与器物特征，画面具有沉浸式历史情境感。",
			TeachingPoints: "文字描述转化为具体图像或场景，增强学生对历史情境的感知",
		},
		{
			Scene: "文学作品可视化", Status: "enabled",
			PromptTemplate: "将文学作品中的情节、场景与人物形象转化为画面：{{text}}。要求：忠于原著意境，构图体现人物关系与情绪张力。",
			TeachingPoints: "情节/场景/人物形象图像化，帮助学生感受作品艺术魅力",
		},
		{
			Scene: "地理地貌展示", Status: "enabled",
			PromptTemplate: "将地理文献中的地貌描述转化为地形示意图：{{text}}。要求：准确呈现地形类型、地势起伏与地貌特征。",
			TeachingPoints: "描述转化为地形图/地貌图，直观促进对地理特征的理解",
		},
		{
			Scene: "图像作文训练", Status: "enabled",
			PromptTemplate: "生成一张用于作文训练的启发图像：{{text}}。要求：画面留有想象空间，适合学生根据图像构思故事或描述场景。",
			TeachingPoints: "图像启发写作，锻炼视觉信息到文字表达的转化能力",
		},
		{
			Scene: "跨学科学习", Status: "enabled",
			PromptTemplate: "将不同学科知识以图像形式呈现：{{text}}。要求：跨学科要素可视化，便于学生将图像转化为文字描述。",
			TeachingPoints: "知识以图像形式呈现，丰富认知体验并促进深入理解",
		},
	}
	for i := range presets {
		var exist model.ArtsciSceneTemplate
		err := global.GVA_DB.WithContext(ctx).Where("scene = ?", presets[i].Scene).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&presets[i]).Error; err != nil {
				global.GVA_LOG.Error("artsci 场景模板种子失败: " + err.Error())
			}
		}
	}
}
