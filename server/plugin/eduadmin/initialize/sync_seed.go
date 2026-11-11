// Package initialize eduadmin：数据同步适配器种子（需求 3.6-29~33，五类来源按名幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// syncAdapterSeed 五类来源适配器（端点为占位配置，凭据留空由管理员录入）
var syncAdapterSeed = []model.EduSyncAdapter{
	{Source: "人事", Protocol: "HTTP", Endpoint: "http://hr-system.example.com/api/v1/staff/sync", SyncFreqMinutes: 5, Status: "停用"},
	{Source: "财务", Protocol: "WebService", Endpoint: "http://fin-system.example.com/ws/budget/sync", SyncFreqMinutes: 5, Status: "停用"},
	{Source: "资产", Protocol: "HTTP", Endpoint: "http://asset-system.example.com/api/v1/assets/sync", SyncFreqMinutes: 5, Status: "停用"},
	{Source: "后勤", Protocol: "HTTP", Endpoint: "http://logistics.example.com/api/v1/dorm/sync", SyncFreqMinutes: 5, Status: "停用"},
	{Source: "教学科研", Protocol: "WebService", Endpoint: "http://academia.example.com/ws/projects/sync", SyncFreqMinutes: 5, Status: "停用"},
}

// SeedEduSyncAdapters 适配器种子（按来源幂等）
func SeedEduSyncAdapters(ctx context.Context) {
	for i := range syncAdapterSeed {
		var exist model.EduSyncAdapter
		err := global.GVA_DB.WithContext(ctx).Where("source = ?", syncAdapterSeed[i].Source).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&syncAdapterSeed[i]).Error; err != nil {
				global.GVA_LOG.Error("eduadmin 同步适配器种子失败: " + err.Error())
			}
		}
	}
}
