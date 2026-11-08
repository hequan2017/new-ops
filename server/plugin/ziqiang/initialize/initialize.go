// Package initialize ziqiang 自强-理工科大模型插件初始化（表/API/菜单/授权/路由，全部幂等）
package initialize

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/middleware"
	sysModel "github.com/hequan2017/new-ops/server/model/system"
	pluginUtils "github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
	srvUtils "github.com/hequan2017/new-ops/server/utils"

	"github.com/hequan2017/new-ops/server/plugin/ziqiang/router"
)

// Gorm 注册 ziqiang 数据表（幂等）。M0 骨架暂无业务表，O02 语料起在此追加 AutoMigrate。
func Gorm(ctx context.Context) {
	_ = ctx
}

// Router 注册 ziqiang 路由（private 组显式挂 JWT+casbin）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}

// Api 注册 ziqiang API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/ziqiang/ping", Description: "自强理工科大模型连通性探测", ApiGroup: "自强理工科大模型", Method: "GET"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 ziqiang 菜单（自强理工科大模型父级，子页面随 O11 挂其下）
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "ziqiang",
			Name:      "ziqiang",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      8,
			Meta:      sysModel.Meta{Title: "自强理工科大模型", Icon: "cpu"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册 ziqiang 策略（幂等）：仅 888 全量
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/ziqiang/ping", "GET"},
	} {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			global.GVA_LOG.Error("ziqiang casbin 添加失败: " + err.Error())
		}
	}
}
