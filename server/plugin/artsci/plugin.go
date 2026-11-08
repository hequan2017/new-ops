// Package artsci 白泽业务插件：文理学科 AI 工具集（数学求解器/作业批改/出题组卷/文生图教学场景）
// 开发规范见 docs/plugin-dev-guide.md；冲刺排期见 docs/dev-plan-20261108.md；错误码 2500-2599
package artsci

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/plugin/artsci/initialize"
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
	initialize.SeedTools(ctx)
	initialize.Api(ctx)
	initialize.Menu(ctx)
	initialize.Casbin(ctx)
	initialize.Router(group)
}
