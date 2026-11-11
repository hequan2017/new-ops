// Package initialize eduadmin：智能分派规则种子（需求 3.6-36，五类规则按关键词幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// dispatchRuleSeed 五类关键词规则
var dispatchRuleSeed = []model.EduDispatchRule{
	{Keywords: "成绩,选课,转专业,学籍,考试", Dept: "教务处", Priority: 10},
	{Keywords: "宿舍,空调,水电,漏水,维修,食堂", Dept: "后勤保障处", Priority: 10},
	{Keywords: "报销,工资,津贴,学费,发票", Dept: "财务处", Priority: 10},
	{Keywords: "网络,账号,密码,系统,门户,邮箱", Dept: "信息化办公室", Priority: 8},
	{Keywords: "档案,人事,合同,离职,入职", Dept: "人事处", Priority: 8},
}

// SeedEduDispatchRules 分派规则种子（按部门幂等）
func SeedEduDispatchRules(ctx context.Context) {
	for i := range dispatchRuleSeed {
		var exist model.EduDispatchRule
		err := global.GVA_DB.WithContext(ctx).Where("dept = ?", dispatchRuleSeed[i].Dept).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&dispatchRuleSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("eduadmin 分派规则种子失败: " + err.Error())
			}
		}
	}
}
