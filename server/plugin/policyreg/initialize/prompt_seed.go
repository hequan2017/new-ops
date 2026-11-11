// Package initialize policyreg：制度问答提示词种子（需求 3.8-8，5 条按模板名幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
)

// promptSeed 制度问答模板 5 条
var promptSeed = []model.PolicyPromptTemplate{
	{Name: "制度问答-标准", Content: "你是高校制度问答助手。请依据参考制度回答「{{question}}」（提问角色：{{role}}），仅引用密级允许的内容。\n参考制度：\n{{refs}}\n输出：①直接结论 ②依据条款名称 ③注意事项。", Enabled: true},
	{Name: "制度问答-对比口径", Content: "你是制度对比助手。请对比参考制度中关于「{{question}}」的不同口径（新旧/校区/单位差异），标注出处。\n参考制度：\n{{refs}}\n输出差异表 + 建议采用口径。", Enabled: true},
	{Name: "制度问答-流程指引", Content: "你是办事流程指引助手。针对「{{question}}」按参考制度给出办理步骤（材料/部门/时限）。\n参考制度：\n{{refs}}\n分步骤输出，最后列出常见退回原因。", Enabled: true},
	{Name: "制度问答-合规检查", Content: "你是合规检查助手。请判断「{{question}}」所述情形是否符合参考制度要求，指出违反条款与风险等级（高/中/低）。\n参考制度：\n{{refs}}\n输出：合规结论 + 违反条款 + 整改建议。", Enabled: true},
	{Name: "制度问答-教学科研专区", Content: "你是教学科研制度助手。请围绕「{{question}}」为「{{role}}」梳理可用政策工具（项目/经费/平台），标注密级与申请条件。\n参考制度：\n{{refs}}\n输出清单式要点。", Enabled: true},
}

// SeedPolicyPrompts 提示词种子（按模板名幂等）
func SeedPolicyPrompts(ctx context.Context) {
	for i := range promptSeed {
		var exist model.PolicyPromptTemplate
		err := global.GVA_DB.WithContext(ctx).Where("name = ?", promptSeed[i].Name).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&promptSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("policyreg 提示词种子失败: " + err.Error())
			}
		}
	}
}
