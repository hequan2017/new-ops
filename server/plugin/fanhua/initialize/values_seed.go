// Package initialize fanhua：价值观知识库种子（需求 3.4-16，按名/词幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// seedWords 敏感词（权重=健康度单次扣减）
var seedWords = []model.FanhuaSensitiveWord{
	{Word: "赌博", Action: "block", Weight: 30, Enabled: true},
	{Word: "诈骗", Action: "block", Weight: 30, Enabled: true},
	{Word: "色情", Action: "block", Weight: 40, Enabled: true},
	{Word: "毒品", Action: "block", Weight: 40, Enabled: true},
	{Word: "洗钱", Action: "block", Weight: 30, Enabled: true},
	{Word: "邪教", Action: "block", Weight: 40, Enabled: true},
	{Word: "私彩", Action: "block", Weight: 25, Enabled: true},
	{Word: "代孕", Action: "block", Weight: 20, Enabled: true},
	{Word: "翻墙软件", Action: "block", Weight: 20, Enabled: true},
	{Word: "迷信偏方", Action: "block", Weight: 15, Enabled: true},
	{Word: "傻子", Action: "replace", Replacement: "反应较慢的人", Weight: 5, Enabled: true},
	{Word: "废物", Action: "replace", Replacement: "表现不佳者", Weight: 5, Enabled: true},
}

// seedTemplates 合规表达模板
var seedTemplates = []model.FanhuaComplianceTemplate{
	{Name: "方言问答合规模板", Scene: "沪语问答", Template: "回答须符合公序良俗，涉方言俚语时给出普通话对照，不输出任何歧视性表述。"},
	{Name: "新闻转述模板", Scene: "舆情转述", Template: "转述新闻须标注来源与时间，只陈述事实不做价值判断，敏感表述以中性语言替换。"},
	{Name: "评论区引导模板", Scene: "社区互动", Template: "引导理性讨论：先复述对方观点，再表达分歧，禁止人身攻击与地域标签。"},
	{Name: "多语审核模板", Scene: "中英混合内容", Template: "中英混排内容按同标准审核，外语俚语与歧视语一并识别，必要时给出双语说明。"},
	{Name: "教育场景模板", Scene: "教学对话", Template: "面向学生输出须积极正向，纠正不当表达时说明原因并给出合规替代表达。"},
}

// seedCases 多语言价值观冲突案例
var seedCases = []model.FanhuaConflictCase{
	{Lang: "中", Content: "有人认为言论自由意味着可以任意辱骂他人。", RiskNote: "言论自由以不侵犯他人权利为边界"},
	{Lang: "英", Content: "Claiming hate speech is just free speech.", RiskNote: "Hate speech falls outside protected speech"},
	{Lang: "中", Content: "以传统习俗为由拒绝女性接受教育。", RiskNote: "受教育权为基本权利，习俗不能豁免"},
	{Lang: "英", Content: "Stereotyping a whole region as dishonest.", RiskNote: "地域歧视性刻板印象"},
	{Lang: "多语", Content: "外语俚语中含贬损群体的称呼并主张其合理性。", RiskNote: "贬损性称呼不因语言切换而合规"},
	{Lang: "中", Content: "声称民间偏方可以替代正规医疗。", RiskNote: "健康信息须符合科学共识"},
	{Lang: "英", Content: "Sharing unverified health cures as facts.", RiskNote: "Unverified medical claims"},
	{Lang: "中", Content: "以群聊名义组织资金盘并称稳赚不赔。", RiskNote: "涉非法集资风险"},
	{Lang: "英", Content: "Promoting gambling as easy income.", RiskNote: "Gambling promotion"},
	{Lang: "中", Content: "对未成年人输出暴力解决问题的内容。", RiskNote: "未成年人保护优先"},
	{Lang: "英", Content: "Encouraging minors to bypass safety rules.", RiskNote: "Minor safety"},
	{Lang: "多语", Content: "多语言社区中对不同方言使用者的贬低言论。", RiskNote: "方言平等，语言歧视不可取"},
}

// SeedValues 价值观知识库种子（词/模板/案例三类，按唯一键幂等）
func SeedValues(ctx context.Context) {
	for i := range seedWords {
		var exist model.FanhuaSensitiveWord
		err := global.GVA_DB.WithContext(ctx).Where("word = ?", seedWords[i].Word).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&seedWords[i]).Error; err != nil {
				global.GVA_LOG.Error("fanhua 敏感词种子失败: " + err.Error())
			}
		}
	}
	for i := range seedTemplates {
		var exist model.FanhuaComplianceTemplate
		err := global.GVA_DB.WithContext(ctx).Where("name = ?", seedTemplates[i].Name).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&seedTemplates[i]).Error; err != nil {
				global.GVA_LOG.Error("fanhua 合规模板种子失败: " + err.Error())
			}
		}
	}
	for i := range seedCases {
		var exist model.FanhuaConflictCase
		err := global.GVA_DB.WithContext(ctx).Where("content = ?", seedCases[i].Content).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&seedCases[i]).Error; err != nil {
				global.GVA_LOG.Error("fanhua 冲突案例种子失败: " + err.Error())
			}
		}
	}
}
