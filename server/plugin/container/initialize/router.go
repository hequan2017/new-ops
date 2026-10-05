// Package initialize 白泽容器管理插件初始化
package initialize

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/global"
	sysModel "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/plugin/container/router"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
)

// Router 注册容器管理路由（private 组，casbin 鉴权）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	router.RouterGroupApp.Init(private)
}

// Api 注册容器管理 API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/container/endpoint", Description: "创建接入点", ApiGroup: "容器管理", Method: "POST"},
		{Path: "/container/endpoint", Description: "更新接入点", ApiGroup: "容器管理", Method: "PUT"},
		{Path: "/container/endpoint", Description: "删除接入点", ApiGroup: "容器管理", Method: "DELETE"},
		{Path: "/container/endpoint/list", Description: "接入点列表", ApiGroup: "容器管理", Method: "GET"},
		{Path: "/container/endpoint/check", Description: "手动巡检", ApiGroup: "容器管理", Method: "POST"},
	}
	utils.RegisterApis(entities...)
}

// Menu 注册容器管理菜单
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "container",
			Name:      "container",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      6,
			Meta:      sysModel.Meta{Title: "容器管理", Icon: "box"},
		},
		{
			Path:      "containerEndpoint",
			Name:      "containerEndpoint",
			Hidden:    false,
			Component: "plugin/container/view/endpoint/index.vue",
			Sort:      1,
			Meta:      sysModel.Meta{Title: "接入点管理", Icon: "guide"},
		},
	}
	utils.RegisterMenus(entities...)
}
