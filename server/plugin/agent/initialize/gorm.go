// Package initialize 白泽 Agent 通道初始化（M9 A1）
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/agent/model"
)

// Gorm 注册 Agent 数据表（幂等）
func Gorm(ctx context.Context) {
	if err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.AgentInstance),
	); err != nil {
		global.GVA_LOG.Error("agent 注册表失败: " + err.Error())
	}
}
