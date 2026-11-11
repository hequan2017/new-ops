// Package initialize crossauth：SSO 应用种子（需求 3.7-4，三应用按 appKey 幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// ssoAppSeed 跨校应用 3 个
var ssoAppSeed = []model.CaSsoApp{
	{AppName: "跨校 AI 开发平台", AppKey: "app-ai-dev", CallbackUrl: "https://ai-dev.university-alliance.edu.cn/sso/callback", Status: "启用"},
	{AppName: "大模型服务器群", AppKey: "app-llm-cluster", CallbackUrl: "https://llm.university-alliance.edu.cn/sso/callback", Status: "启用"},
	{AppName: "跨校学科工具集", AppKey: "app-toolset", CallbackUrl: "https://tools.university-alliance.edu.cn/sso/callback", Status: "启用"},
}

// SeedCaSsoApps SSO 应用种子（按 appKey 幂等）
func SeedCaSsoApps(ctx context.Context) {
	for i := range ssoAppSeed {
		var exist model.CaSsoApp
		err := global.GVA_DB.WithContext(ctx).Where("app_key = ?", ssoAppSeed[i].AppKey).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&ssoAppSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("crossauth SSO 应用种子失败: " + err.Error())
			}
		}
	}
}
