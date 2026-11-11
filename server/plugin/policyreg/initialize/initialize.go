// Package initialize policyreg 高校制度大模型插件初始化（表/API/菜单/授权/路由，全部幂等）
package initialize

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/middleware"
	sysModel "github.com/hequan2017/new-ops/server/model/system"
	pluginUtils "github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
	srvUtils "github.com/hequan2017/new-ops/server/utils"

	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
	"github.com/hequan2017/new-ops/server/plugin/policyreg/router"
)

// Gorm 注册 policyreg 数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.PolicyDoc),
	)
	if err != nil {
		err = errors.Wrap(err, "policyreg 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册 policyreg 路由（private 组显式挂 JWT+casbin）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}

// Api 注册 policyreg API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/policyreg/ping", Description: "高校制度大模型连通性探测", ApiGroup: "高校制度大模型", Method: "GET"},
		{Path: "/policyreg/doc/list", Description: "制度列表", ApiGroup: "高校制度大模型", Method: "GET"},
		{Path: "/policyreg/doc", Description: "制度详情", ApiGroup: "高校制度大模型", Method: "GET"},
		{Path: "/policyreg/doc", Description: "导入制度", ApiGroup: "高校制度大模型", Method: "POST"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 policyreg 菜单（高校制度大模型父级，子页面随 R09 挂其下）
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "policyreg",
			Name:      "policyreg",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      11,
			Meta:      sysModel.Meta{Title: "高校制度大模型", Icon: "document"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册 policyreg 策略（幂等）：仅 888 全量
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/policyreg/ping", "GET"},
		{"/policyreg/doc/list", "GET"},
		{"/policyreg/doc", "GET"},
		{"/policyreg/doc", "POST"},
	} {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			global.GVA_LOG.Error("policyreg casbin 添加失败: " + err.Error())
		}
	}
}
