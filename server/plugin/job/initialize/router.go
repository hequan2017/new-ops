// Package initialize 白泽批量作业插件初始化
package initialize

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/middleware"
	"github.com/hequan2017/new-ops/server/plugin/job/router"
)

// Router 注册批量作业路由（private 组显式挂 JWT+casbin——插件自建组不继承底座中间件）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}
