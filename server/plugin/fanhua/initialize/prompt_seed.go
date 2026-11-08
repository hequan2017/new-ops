// Package initialize fanhua：提示词库种子（需求 3.4-18/19，按名幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// promptSeed 沪语场景 8 + 价值观审核链式 3
var promptSeed = []model.FanhuaPromptTemplate{
	{Name: "沪语解释", Scene: "沪语场景", Content: "用上海话解释下面的内容，并给出普通话对照：{{text}}", Enabled: true},
	{Name: "沪语翻译", Scene: "沪语场景", Content: "把下面普通话句子翻译成上海话：{{text}}", Enabled: true},
	{Name: "沪语故事续写", Scene: "沪语场景", Content: "以沪语续写下面故事的后续（300 字内），结尾给出普通话大意：{{text}}", Enabled: true},
	{Name: "弄堂场景对话", Scene: "沪语场景", Content: "生成一段发生在{{scene}}的沪语生活对话（2-3 人），附普通话注释。", Enabled: true},
	{Name: "沪语童谣讲解", Scene: "沪语场景", Content: "讲解这首沪语童谣的出处、玩法与文化含义：{{rhyme}}", Enabled: true},
	{Name: "沪剧选段赏析", Scene: "沪语场景", Content: "赏析以下沪剧选段的唱词特色与情感表达：{{excerpt}}", Enabled: true},
	{Name: "海派知识问答", Scene: "沪语场景", Content: "以海派文化专家口吻回答：{{question}}（沪语词汇首次出现给普通话对照）", Enabled: true},
	{Name: "沪语俚语建议", Scene: "沪语场景", Content: "说明沪语词「{{slang}}」的含义、使用场合与例句。", Enabled: true},
	{Name: "价值观审核-检测", Scene: "价值观审核", Content: "对以下文本做价值观风险检测，只输出 JSON {\"risk\":\"高|中|低\",\"hits\":[\"命中项\"]}：{{text}}", Enabled: true},
	{Name: "价值观审核-修正", Scene: "价值观审核", Content: "将以下文本改写为合规表达，保持原意与语气，不引入新信息：{{text}}", Enabled: true},
	{Name: "价值观审核-复核", Scene: "价值观审核", Content: "复核以下改写后文本是否仍符合规范，输出 JSON {\"pass\":true|false,\"reason\":\"...\"}：{{text}}", Enabled: true},
}

// SeedPrompts 提示词库种子（按模板名幂等）
func SeedPrompts(ctx context.Context) {
	for i := range promptSeed {
		var exist model.FanhuaPromptTemplate
		err := global.GVA_DB.WithContext(ctx).Where("name = ?", promptSeed[i].Name).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&promptSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("fanhua 提示词种子失败: " + err.Error())
			}
		}
	}
}
