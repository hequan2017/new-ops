// Package initialize crossauth：身份互认规则种子（需求 3.7-2，8 条按身份+资源幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// trustRuleSeed 互认规则 8 条
var trustRuleSeed = []model.CaTrustRule{
	{HomeIdentity: "上海大学教师", TargetResource: "合作高校教学资源库", Scope: "课程课件,教学视频", Allowed: true, Remark: "教学协作互认"},
	{HomeIdentity: "上海大学教师", TargetResource: "合作高校图书馆", Scope: "电子书库,文献传递", Allowed: true, Remark: "图书馆联盟"},
	{HomeIdentity: "上海大学学生", TargetResource: "合作高校教学资源库", Scope: "课程课件", Allowed: true, Remark: "学分互认配套"},
	{HomeIdentity: "上海大学科研人员", TargetResource: "合作高校算力平台", Scope: "GPU 队列,存储配额", Allowed: true, Remark: "科研算力共享"},
	{HomeIdentity: "上海大学学生", TargetResource: "合作高校算力平台", Scope: "教学实验算力", Allowed: false, Remark: "学生不对等开放"},
	{HomeIdentity: "合作高校教师", TargetResource: "上海大学实验室平台", Scope: "预约,设备清单", Allowed: true, Remark: "反向互认"},
	{HomeIdentity: "企业用户", TargetResource: "合作高校教学资源库", Scope: "公开课", Allowed: true, Remark: "产学合作"},
	{HomeIdentity: "企业用户", TargetResource: "上海大学实验室平台", Scope: "", Allowed: false, Remark: "未签约企业默认拒绝"},
}

// SeedCaTrustRules 互认规则种子（按 身份+资源 幂等）
func SeedCaTrustRules(ctx context.Context) {
	for i := range trustRuleSeed {
		var exist model.CaTrustRule
		err := global.GVA_DB.WithContext(ctx).
			Where("home_identity = ? AND target_resource = ?", trustRuleSeed[i].HomeIdentity, trustRuleSeed[i].TargetResource).
			First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&trustRuleSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("crossauth 互认规则种子失败: " + err.Error())
			}
		}
	}
}
