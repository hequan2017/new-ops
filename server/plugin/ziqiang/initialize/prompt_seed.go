// Package initialize ziqiang：侵权检测提示词种子（需求 3.5-12/13，按模板名幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// promptSeed 多模态双验证 1 + 引导式对话三步 3
var promptSeed = []model.ZqPromptTemplate{
	{
		Name:  "侵权检测-多模态双验证",
		Scene: "多模态双验证", Step: 1,
		Content: "你是知识产权侵权鉴定专家。请对疑似侵权对象执行多模态双验证并输出结论：\n" +
			"一、文本相似度验证：待检文本片段「{{textSnippet}}」与比对文本「{{referenceText}}」的滑窗 bigram 相似度为 {{textSimilarity}}（阈值 {{textThreshold}}），结合法条术语命中情况判定文本维度是否构成实质相似。\n" +
			"二、图像特征验证：待检图像 aHash={{imageHashA}}，比对图像 aHash={{imageHashB}}，汉明距离 {{hammingDistance}}（阈值 {{hammingThreshold}}），结合水印标记与设备指纹判定图像维度是否构成近似复制。\n" +
			"三、综合结论：两维度独立判定后给出「文本成立/图像成立/双维成立/均不成立」结论，并列出证据要点与剩余风险。",
		Vars: "", Enabled: true,
	},
	{
		Name:  "引导式对话-第一步创作场景",
		Scene: "引导式对话", Step: 1,
		Content: "你是知识产权引导式咨询助手（第一步/共三步）。请围绕用户的创作场景采集要素：\n" +
			"创作场景描述：{{scene}}\n" +
			"请输出：①创作客体类型（文字/图像/模型/代码）②完成时间与首次公开方式 ③是否已登记或发表 ④引导用户补充创作过程证据（草稿、时间戳、设备记录）。输出需简洁分点，供下一步确认权利类型使用。",
		Vars: "", Enabled: true,
	},
	{
		Name:  "引导式对话-第二步权利类型",
		Scene: "引导式对话", Step: 2,
		Content: "你是知识产权引导式咨询助手（第二步/共三步）。基于第一步场景要素「{{sceneDigest}}」，请确认权利类型：\n" +
			"用户意向权利类型：{{rightsType}}\n" +
			"请输出：①该类型对应的保护范围与期限（专利/著作权/商标/商业秘密口径）②与场景要素的匹配度分析 ③若权利类型与场景不符给出更优选择与理由 ④引导用户进入第三步法律文书准备所需材料清单。",
		Vars: "", Enabled: true,
	},
	{
		Name:  "引导式对话-第三步法律文书",
		Scene: "引导式对话", Step: 3,
		Content: "你是知识产权引导式咨询助手（第三步/共三步）。基于前两步确认的场景「{{sceneDigest}}」与权利类型「{{rightsType}}」，结合对方行为「{{infringingAct}}」，请生成法律文书初稿：\n" +
			"文书类型：{{docType}}（侵权告知函/投诉书/起诉状初稿 三选一）\n" +
			"请输出：①权利基础与归属陈述 ②侵权行为事实与证据清单 ③法律依据条款（引用模型鉴权知识库法规名）④诉求与限期回应要求 ⑤免责声明（本稿为初稿，须经执业律师复核）。文书口吻正式，事实部分用占位符保留待补充项。",
		Vars: "", Enabled: true,
	},
}

// SeedPrompts 提示词种子（按模板名幂等）
func SeedPrompts(ctx context.Context) {
	for i := range promptSeed {
		var exist model.ZqPromptTemplate
		err := global.GVA_DB.WithContext(ctx).Where("name = ?", promptSeed[i].Name).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&promptSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("ziqiang 提示词种子失败: " + err.Error())
			}
		}
	}
}
