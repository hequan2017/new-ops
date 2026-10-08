// Package agent 白泽业务插件：Agent 通道（M9：反向长连接/采集上报/执行代理）
// A1 范围：注册/心跳/系统指标采集上报，资产页展示在线状态；指标落 monitor_metric。
package agent

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/plugin/agent/initialize"
	interfaces "github.com/hequan2017/new-ops/server/utils/plugin/v2"
)

var _ interfaces.Plugin = (*plugin)(nil)

var Plugin = new(plugin)

type plugin struct{}

func init() {
	interfaces.Register(Plugin)
}

// Register 插件加载时调用：表迁移 → API/授权种子 → 路由（REST private + WS 裸注册）
func (p *plugin) Register(group *gin.Engine) {
	ctx := context.Background()
	initialize.Gorm(ctx)
	initialize.Api(ctx)
	initialize.Casbin(ctx)
	initialize.Router(group)
}
