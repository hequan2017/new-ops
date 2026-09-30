// Package pipeline 白泽业务插件：流水线与发布（Pipeline-Stage-Step）
// 开发规范见 docs/DEV_PLAN.md 与 docs/plugin-dev-guide.md
package pipeline

import (
	"github.com/gin-gonic/gin"

	interfaces "github.com/hequan2017/new-ops/server/utils/plugin/v2"
)

var _ interfaces.Plugin = (*plugin)(nil)

var Plugin = new(plugin)

type plugin struct{}

func init() {
	interfaces.Register(Plugin)
}

// Register 插件加载时调用；M3 起实现路由/菜单/API/字典初始化。
func (p *plugin) Register(group *gin.Engine) {
}
