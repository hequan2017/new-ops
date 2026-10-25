// Package dbops 白泽业务插件：数据库工单（实例纳管/SQL 审核）
// 开发规范见 docs/plugin-dev-guide.md
package dbops

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/plugin/dbops/initialize"
	"github.com/hequan2017/new-ops/server/plugin/dbops/service"
	interfaces "github.com/hequan2017/new-ops/server/utils/plugin/v2"
)

var _ interfaces.Plugin = (*plugin)(nil)

var Plugin = new(plugin)

type plugin struct{}

func init() {
	interfaces.Register(Plugin)
}

// Register 插件加载时调用：表迁移 → API/菜单种子 → 授权种子 → 路由
func (p *plugin) Register(group *gin.Engine) {
	ctx := context.Background()
	initialize.Gorm(ctx)
	initialize.Api(ctx)
	initialize.Menu(ctx)
	initialize.Casbin(ctx)
	service.DbopsServiceI.SweepStuckExecuting() // 清扫上次运行滞留的"执行中"工单（无出口状态）
	initialize.Router(group)
}
