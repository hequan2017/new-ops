// Package fanhua 白泽业务插件：繁花-文科大模型（沪语语料/标注/知识库/价值观审查/城市科学/问答 API）
// 开发规范见 docs/plugin-dev-guide.md；冲刺排期见 docs/dev-plan-20261108.md；错误码 2600-2699
package fanhua

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/initialize"
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
	initialize.SeedParallels(ctx)
	initialize.SeedEntities(ctx)
	initialize.SeedVoiceEntries(ctx)
	initialize.SeedValues(ctx)
	initialize.Api(ctx)
	initialize.Menu(ctx)
	initialize.Casbin(ctx)
	initialize.Router(group)
}
