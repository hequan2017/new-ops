// Package initialize ziqiang：模型鉴权知识库种子（需求 3.5-7，15 法规+6 判例，按名幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// regulationSeed 法规种子（中国 6 / 美国 5 / 欧盟 4）
var regulationSeed = []model.ZqRegulation{
	{Jurisdiction: "中国", Category: "版权", Title: "中华人民共和国著作权法", Content: "作品自创作完成即享有著作权；符合作品要件的 AI 生成内容依司法实践个案认定；保护期为作者终生及死后 50 年。", EffectiveDate: "2021-06-01"},
	{Jurisdiction: "中国", Category: "专利", Title: "中华人民共和国专利法", Content: "发明 20 年、实用新型 10 年、外观设计 15 年；保护范围以权利要求内容为准；2021 年修法引入惩罚性赔偿。", EffectiveDate: "2021-06-01"},
	{Jurisdiction: "中国", Category: "商标", Title: "中华人民共和国商标法", Content: "注册商标专用权以核准注册的商标和核定使用的商品为限；恶意侵权可适用惩罚性赔偿。", EffectiveDate: "2019-11-01"},
	{Jurisdiction: "中国", Category: "AI治理", Title: "生成式人工智能服务管理暂行办法", Content: "生成式 AI 服务提供者应承担网络信息内容生产者责任，训练数据合法性、内容标识与安全评估义务。", EffectiveDate: "2023-08-15"},
	{Jurisdiction: "中国", Category: "数据合规", Title: "中华人民共和国数据安全法", Content: "数据分类分级保护制度；重要数据出境须安全评估；数据处理活动全流程安全义务。", EffectiveDate: "2021-09-01"},
	{Jurisdiction: "中国", Category: "数据合规", Title: "中华人民共和国个人信息保护法", Content: "个人信息处理须合法正当必要诚信；敏感个人信息单独同意；自动化决策透明度义务。", EffectiveDate: "2021-11-01"},
	{Jurisdiction: "美国", Category: "版权", Title: "Copyright Act (17 U.S.C.)", Content: "Original works of authorship protection; USCO guidance: human authorship required, purely AI-generated content not registrable.", EffectiveDate: "1976-10-19"},
	{Jurisdiction: "美国", Category: "专利", Title: "Patent Act (35 U.S.C.)", Content: "Patentable subject matter, novelty and non-obviousness; Ai-generated inventions lack inventor status per Thaler doctrine.", EffectiveDate: "1952-01-01"},
	{Jurisdiction: "美国", Category: "版权", Title: "DMCA", Content: "Safe harbor for online service providers; notice-and-takedown; anti-circumvention provisions relevant to model watermarking.", EffectiveDate: "2000-10-28"},
	{Jurisdiction: "美国", Category: "AI治理", Title: "Executive Order 14110 on Safe, Secure AI", Content: "AI safety testing, content authentication standards, and federal agency reporting duties for dual-use foundation models.", EffectiveDate: "2023-10-30"},
	{Jurisdiction: "美国", Category: "数据合规", Title: "State Privacy Laws (CCPA/CPRA 为代表)", Content: "California consumer privacy rights; opt-out of automated decision-making; training-data consumer rights interface.", EffectiveDate: "2020-01-01"},
	{Jurisdiction: "欧盟", Category: "AI治理", Title: "EU AI Act", Content: "Risk-based framework: prohibited practices, high-risk obligations, GPAI transparency including training-data summaries.", EffectiveDate: "2024-08-01"},
	{Jurisdiction: "欧盟", Category: "数据合规", Title: "GDPR", Content: "Lawful basis for personal data processing; data minimisation; automated decision-making Art.22 rights; DPIA duty.", EffectiveDate: "2018-05-25"},
	{Jurisdiction: "欧盟", Category: "版权", Title: "DSM Copyright Directive", Content: "Text and data mining exceptions (Art.3/4); rights holders opt-out for TDM of lawfully accessible works.", EffectiveDate: "2021-06-07"},
	{Jurisdiction: "欧盟", Category: "商标", Title: "EU Trade Mark Regulation", Content: "EUTM unitary registration; likelihood-of-confusion test; well-known mark protection.", EffectiveDate: "2017-10-01"},
}

// caseSeed 判例种子（6 条）
var caseSeed = []model.ZqCase{
	{Jurisdiction: "中国", Name: "腾讯 Dreamwriter 案（深圳南山法院）", Summary: "AI 辅助生成的财经文章被抄袭，腾讯主张著作权。", Ruling: "认定涉案文章体现团队个性化选择与表达，构成法人作品受著作权保护。", Year: 2019},
	{Jurisdiction: "中国", Name: "AI 文生图著作权第一案（北京互联网法院）", Summary: "用户以 Stable Diffusion 生成图片被使用，主张著作权。", Ruling: "认定用户通过提示词与参数调整体现智力投入，图片具备独创性，享有著作权。", Year: 2023},
	{Jurisdiction: "美国", Name: "Thaler v. Perlmutter", Summary: "Thaler 将 AI 列为发明人与作者申请专利与版权登记被拒。", Ruling: "确认人类作者身份为版权登记要件，纯 AI 生成内容不予登记。", Year: 2023},
	{Jurisdiction: "美国", Name: "Google v. Oracle America", Summary: "Java API 调用结构是否构成合理使用的十年争议。", Ruling: "最高法院认定 Google 复制 API 声明构成 transformative 合理使用。", Year: 2021},
	{Jurisdiction: "欧盟", Name: "CJEU Reprobel 案", Summary: "复印补偿金制度下作者与出版者权益分配争议。", Ruling: "明确补偿金应主要归属作者，出版者举证技术关联。", Year: 2015},
	{Jurisdiction: "中国", Name: "短视频模板案（杭州互联网法院）", Summary: "短视频模板是否构成类电作品的争议。", Ruling: "认定具备独创性表达，构成类电作品，侵权成立。", Year: 2020},
}

// SeedKb 模型鉴权知识库种子（法规按名、判例按案名幂等）
func SeedKb(ctx context.Context) {
	for i := range regulationSeed {
		var exist model.ZqRegulation
		err := global.GVA_DB.WithContext(ctx).Where("title = ?", regulationSeed[i].Title).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&regulationSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("ziqiang 法规种子失败: " + err.Error())
			}
		}
	}
	for i := range caseSeed {
		var exist model.ZqCase
		err := global.GVA_DB.WithContext(ctx).Where("name = ?", caseSeed[i].Name).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&caseSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("ziqiang 判例种子失败: " + err.Error())
			}
		}
	}
}
