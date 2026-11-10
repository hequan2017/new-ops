// Package eduadmin 白泽业务插件：教育管理一模通管大模型（全场景数据接入/六大知识库/智能体工具台/AI-12345 工单）
// 开发规范见 docs/plugin-dev-guide.md；冲刺排期见 docs/dev-plan-20261108.md；错误码 2800-2899
package eduadmin

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/initialize"
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
	initialize.SeedEduPolicies(ctx)
	initialize.Api(ctx)
	initialize.Menu(ctx)
	initialize.Casbin(ctx)
	initialize.Router(group)
}
