// Package initialize pipeline 插件定时任务注册（启动时全量恢复 cron 触发）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/pipeline/model"
	"github.com/hequan2017/new-ops/server/plugin/pipeline/service"
)

// Timer 启动时为所有启用 cron 的流水线注册定时触发（服务重启后恢复）
func Timer(ctx context.Context) {
	var list []model.Pipeline
	if err := global.GVA_DB.WithContext(ctx).
		Where("cron_enabled = ? AND enabled = ?", true, true).Find(&list).Error; err != nil {
		return
	}
	svc := new(service.PipelineService)
	for i := range list {
		svc.SyncPipelineCron(&list[i])
	}
}
