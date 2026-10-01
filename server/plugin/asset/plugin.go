// Package asset 白泽业务插件：资产中心（主机资产/机房机柜/产品线/凭据保险库）
// 开发规范见 docs/DEV_PLAN.md 与 docs/plugin-dev-guide.md
package asset

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/plugin/asset/initialize"
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
	initialize.Router(group)
}
