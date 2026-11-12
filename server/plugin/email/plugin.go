// Package email 白泽邮件插件（v2 接口化注册：原 v1 utils/plugin 机制迁移。
// SMTP 配置直读底座 config.yaml 的 email 段；API 记录与菜单由底座 source/system
// 种子提供；注册路径与中间件链与迁移前一致）
package email

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/plugin/email/initialize"
	interfaces "github.com/hequan2017/new-ops/server/utils/plugin/v2"
)

var _ interfaces.Plugin = (*plugin)(nil)

var Plugin = new(plugin)

type plugin struct{}

func init() {
	interfaces.Register(Plugin)
}

func (p *plugin) Register(group *gin.Engine) {
	initialize.Router(group)
}
