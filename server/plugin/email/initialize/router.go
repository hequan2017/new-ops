// Package initialize 白泽邮件插件初始化
package initialize

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/middleware"
	"github.com/hequan2017/new-ops/server/plugin/email/router"
)

// Router 注册邮件路由（private 组显式挂 JWT+casbin——插件自建组不继承底座中间件；
// 与 v1 时期挂底座 private 组的中间件链等价）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.InitEmailRouter(private)
}
