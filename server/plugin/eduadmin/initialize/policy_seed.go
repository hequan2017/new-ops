// Package initialize eduadmin：六大知识库种子（需求 3.6-4，教育管理指标库+域条目，按标题幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// policySeed 六大知识库种子（决策域指标 3 + 域条目 6）
var policySeed = []model.EduPolicy{
	{Domain: "决策", Title: "指标库-教学质量", Tags: "指标,教学质量", Content: "教学质量指标：生师比、教授为本科生授课比例、课程达成度评价值、学生评教均值、毕业率与学位授予率；目标区间 生师比≤18、评教均值≥4.2/5.0。"},
	{Domain: "决策", Title: "指标库-科研产出", Tags: "指标,科研产出", Content: "科研产出指标：年度论文数（高水平期刊占比）、纵向项目数与经费、横向合作经费、专利授权量与转化率、成果获奖数；目标 高水平占比≥40%、专利转化率≥8%。"},
	{Domain: "决策", Title: "指标库-资金使用效率", Tags: "指标,资金效率", Content: "资金使用效率指标：预算执行率、生均培养成本、设备利用率、专项经费按期支出率、审计问题整改闭环率；目标 预算执行率≥95%、设备利用率≥75%。"},
	{Domain: "人事", Title: "人事-编制与招聘规程", Tags: "人事,招聘", Content: "教师岗位招聘须经编制核定、公开发布、资格审查、试讲答辩、集体决策五步；高级岗位由校聘委员会评议。"},
	{Domain: "财务", Title: "财务-预算调整流程", Tags: "财务,预算", Content: "预算调整须提交调整申请与影响测算，单科目超 10% 报分管校领导审批，超 30% 报校长办公会。"},
	{Domain: "资产", Title: "资产-设备采购与验收", Tags: "资产,采购", Content: "单价 10 万元以上设备采购须三方比价并组织验收小组；验收合格后 5 个工作日内入固定资产账。"},
	{Domain: "后勤", Title: "后勤-食堂安全检查", Tags: "后勤,食堂", Content: "食堂实行日检查、周汇总、月评估；留样 48 小时、晨检记录齐全；未达标窗口限期整改并复检。"},
	{Domain: "跨校", Title: "跨校-资源共建共享办法", Tags: "跨校,共建", Content: "12 所共建高校实行课程互选、学分互认、设备共享；使用跨校资源须经本校教务与对方平台双重备案。"},
}

// SeedEduPolicies 六大知识库种子（按标题幂等）
func SeedEduPolicies(ctx context.Context) {
	for i := range policySeed {
		var exist model.EduPolicy
		err := global.GVA_DB.WithContext(ctx).Where("title = ?", policySeed[i].Title).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&policySeed[i]).Error; err != nil {
				global.GVA_LOG.Error("eduadmin 知识库种子失败: " + err.Error())
			}
		}
	}
}
