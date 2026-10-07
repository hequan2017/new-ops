// Package initialize container 插件表迁移与巡检循环（幂等）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/container/model"
	"github.com/hequan2017/new-ops/server/plugin/container/service"
)

// Gorm 注册容器管理数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.DockerEndpoint),
		new(model.DockerEventLog),
	)
	if err != nil {
		global.GVA_LOG.Error("container 注册表失败: " + err.Error())
	}
}

// Inspect 启动 30s 合并巡检循环（进程生命周期，tianqi 合并巡检模式）
func Inspect(ctx context.Context) {
	service.Service.Endpoint.StartInspectLoop(ctx)
}
