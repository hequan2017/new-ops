// Package initialize eduadmin：跨校协同种子（需求 3.6-27/28，培训材料+最佳实践，按标题幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// practiceSeed 培训材料 3 + 最佳实践 3
var practiceSeed = []model.EduBestPractice{
	{Title: "跨校课程互选操作手册", Category: "培训材料", Content: "课程互选五步：本校教务备案→平台选课→对方高校确认→学分认定→成绩回传；每学期选课上限 2 门。"},
	{Title: "大型仪器共享预约指南", Category: "培训材料", Content: "预约制共享：平台提交使用申请→归属校审核→机时排期→使用登记→耗材结算；培训合格方可独立上机。"},
	{Title: "跨校联合培养导师工作指引", Category: "培训材料", Content: "双导师制：主导师在学籍校、副导师在培养校；每学期至少 2 次联合指导记录，论文双方署名规范。"},
	{Title: "12 校学分互认最佳实践", Category: "最佳实践", Content: "以课程大纲比对替代课程名比对；核心课互认清单每年更新一次，争议课程由教务联席会裁定。"},
	{Title: "实验设备跨校调度最佳实践", Category: "最佳实践", Content: "贵重设备共享采用机时积分制：贡献机时与使用机时年度结算，赤字校次年优先保障排期。"},
	{Title: "跨校数据安全协同要点", Category: "最佳实践", Content: "共享数据按最小必要原则脱敏传输；跨校接口凭据密文保管、双人复核，泄露事件 30 分钟内双方通报。"},
}

// SeedEduPractices 跨校协同种子（按标题幂等）
func SeedEduPractices(ctx context.Context) {
	for i := range practiceSeed {
		var exist model.EduBestPractice
		err := global.GVA_DB.WithContext(ctx).Where("title = ?", practiceSeed[i].Title).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&practiceSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("eduadmin 跨校协同种子失败: " + err.Error())
			}
		}
	}
}
