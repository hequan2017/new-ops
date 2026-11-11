// Package crossauth 白泽业务插件：跨校访问认证与安全服务管理平台（联邦身份/互认/MFA/SSO/权限矩阵/分级/SM4/监控）
// 开发规范见 docs/plugin-dev-guide.md；冲刺排期见 docs/dev-plan-20261108.md；错误码 2900-2999
package crossauth

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/initialize"
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
	initialize.SeedCaTrustRules(ctx)
	initialize.SeedCaSsoApps(ctx)
	initialize.SeedCaRoleTemplates(ctx)
	initialize.Api(ctx)
	initialize.Menu(ctx)
	initialize.Casbin(ctx)
	initialize.Router(group)
}
