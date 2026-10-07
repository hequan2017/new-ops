// Package container 白泽业务插件：容器管理（Docker 接入点/容器/镜像）
// 开发规范见 docs/DEV_PLAN.md 与 docs/plugin-dev-guide.md
package container

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/plugin/container/initialize"
	interfaces "github.com/hequan2017/new-ops/server/utils/plugin/v2"
)

var _ interfaces.Plugin = (*plugin)(nil)

var Plugin = new(plugin)

type plugin struct{}

func init() {
	interfaces.Register(Plugin)
}

// Register 插件加载时调用：表迁移 → API/菜单种子 → 授权种子 → 路由 → 30s 巡检循环
func (p *plugin) Register(group *gin.Engine) {
	ctx := context.Background()
	initialize.Gorm(ctx)
	initialize.Api(ctx)
	initialize.Menu(ctx)
	initialize.Casbin(ctx)
	initialize.Router(group)
	initialize.Inspect(ctx)
}
