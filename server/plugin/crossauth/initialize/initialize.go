// Package initialize crossauth 跨校认证与安全插件初始化（表/API/菜单/授权/路由，全部幂等）
package initialize

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/middleware"
	sysModel "github.com/hequan2017/new-ops/server/model/system"
	pluginUtils "github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
	srvUtils "github.com/hequan2017/new-ops/server/utils"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/router"
)

// Gorm 注册 crossauth 数据表（幂等）。M0 骨架暂无业务表，Q02 联邦身份起在此追加 AutoMigrate。
func Gorm(ctx context.Context) {
	_ = ctx
}

// Router 注册 crossauth 路由（private 组显式挂 JWT+casbin）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}

// Api 注册 crossauth API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/crossauth/ping", Description: "跨校认证与安全连通性探测", ApiGroup: "跨校认证与安全", Method: "GET"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 crossauth 菜单（跨校认证与安全父级，子页面随 Q19 挂其下）
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "crossauth",
			Name:      "crossauth",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      10,
			Meta:      sysModel.Meta{Title: "跨校认证与安全", Icon: "lock"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册 crossauth 策略（幂等）：仅 888 全量
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/crossauth/ping", "GET"},
	} {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			global.GVA_LOG.Error("crossauth casbin 添加失败: " + err.Error())
		}
	}
}
