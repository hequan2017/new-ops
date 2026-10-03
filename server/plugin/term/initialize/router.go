// Package initialize 白泽终端插件路由
package initialize

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/term/router"
)

// Router 注册终端路由（WS 挂 public：握手自验 token）
func Router(engine *gin.Engine) {
	public := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	router.RouterGroupApp.Init(public, private)
}
