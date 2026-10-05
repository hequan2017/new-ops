// Package workflow 白泽业务插件：工单引擎（状态机/审批）
// 开发规范见 docs/DEV_PLAN.md 与 docs/plugin-dev-guide.md
package workflow

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/plugin/workflow/initialize"
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
