// Package oplib 白泽业务插件：算子库（算子注册/执行/编排）
// 开发规范见 docs/plugin-dev-guide.md；冲刺排期见 docs/dev-plan-llm-20261106.md；错误码 2300-2399
package oplib

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/plugin/oplib/initialize"
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
