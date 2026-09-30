// Package job 白泽业务插件：批量作业（命令/脚本/变量组）
// 开发规范见 docs/DEV_PLAN.md 与 docs/plugin-dev-guide.md
package job

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

// Register 插件加载时调用；M1 起实现路由/菜单/API/字典初始化。
func (p *plugin) Register(group *gin.Engine) {
}
