// Package initialize oplib 算子库插件初始化（表/API/菜单/授权/路由，全部幂等）
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

	"github.com/hequan2017/new-ops/server/plugin/oplib/model"
	"github.com/hequan2017/new-ops/server/plugin/oplib/router"
)

// Gorm 注册 oplib 数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.OplibOperator),
		new(model.OplibRunLog),
		new(model.OplibPipeline),
	)
	if err != nil {
		err = errors.Wrap(err, "oplib 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册 oplib 路由（private 组显式挂 JWT+casbin）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}

// Api 注册 oplib API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/oplib/operator/list", Description: "算子列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/oplib/pipeline/list", Description: "算子管道列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/oplib/pipeline", Description: "保存算子管道", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/oplib/pipeline/run", Description: "执行算子管道", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/oplib/pipeline", Description: "删除算子管道", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/oplib/operator", Description: "保存算子", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/oplib/operator/run", Description: "执行算子", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/oplib/operator/runLogs", Description: "算子执行留痕", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/oplib/operator", Description: "删除算子", ApiGroup: "大模型平台", Method: "DELETE"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 oplib 菜单（首元素为大模型平台父级引用，子页面挂其下）
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "llmops",
			Name:      "llmops",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      5,
			Meta:      sysModel.Meta{Title: "大模型平台", Icon: "data-line"},
		},
		{
			Path:      "oplibOperator",
			Name:      "oplibOperator",
			Hidden:    false,
			Component: "plugin/oplib/view/operator/index.vue",
			Sort:      8,
			Meta:      sysModel.Meta{Title: "算子库", Icon: "set-up"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册 oplib 策略（幂等）：仅 888 全量
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/oplib/operator/list", "GET"},
		{"/oplib/pipeline/list", "GET"},
		{"/oplib/pipeline", "POST"},
		{"/oplib/pipeline/run", "POST"},
		{"/oplib/pipeline", "DELETE"},
		{"/oplib/operator", "POST"},
		{"/oplib/operator/run", "POST"},
		{"/oplib/operator/runLogs", "GET"},
		{"/oplib/operator", "DELETE"},
	} {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			global.GVA_LOG.Error("oplib casbin 添加失败: " + err.Error())
		}
	}
}
