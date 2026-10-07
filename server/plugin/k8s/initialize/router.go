// Package initialize k8s 插件路由
package initialize

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/middleware"
	"github.com/hequan2017/new-ops/server/plugin/k8s/router"
)

// Router 注册 k8s 路由（private 组显式挂 JWT+casbin——插件自建组不继承底座中间件；
// public 组留给 WS 流端点，握手时自验 query token）
func Router(engine *gin.Engine) {
	public := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(public, private)
}
