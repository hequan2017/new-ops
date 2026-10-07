// Package initialize 白泽流水线插件初始化
package initialize

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/middleware"
	"github.com/hequan2017/new-ops/server/plugin/pipeline/router"
)

// Router 注册流水线路由（private 组显式挂 JWT+casbin——插件自建组不继承底座中间件；
// SSE/webhook 挂 public 自验 query token/路径令牌）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	public := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	router.RouterGroupApp.Init(private, public)
}
