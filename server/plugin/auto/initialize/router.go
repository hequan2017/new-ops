package initialize

import (
	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/middleware"
	"github.com/hequan2017/new-ops/server/plugin/auto/router"
)

func Router(engine *gin.Engine) {
	InitializeRouter(engine)
}

func InitializeRouter(engine *gin.Engine) {
	public := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())

	router.RouterGroupApp.InitAutoCodeRouter(private, public)
	router.RouterGroupApp.InitAutoCodeHistoryRouter(private)
	router.RouterGroupApp.InitSkillsRouter(private, public)
}
