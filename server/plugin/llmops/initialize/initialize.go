// Package initialize llmops 大模型应用开发平台插件初始化（表/API/菜单/授权/路由，全部幂等）
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

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
	"github.com/hequan2017/new-ops/server/plugin/llmops/router"
)

// Gorm 注册 llmops 数据表（幂等）；表随排期条目逐域新增，新增在此追加
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.LlmopsDataAsset),
	)
	if err != nil {
		err = errors.Wrap(err, "llmops 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册 llmops 路由（private 组显式挂 JWT+casbin）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}

// Api 注册 llmops API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/llmops/ping", Description: "llmops 连通性探测", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/asset/list", Description: "数据资产列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/asset", Description: "数据资产详情", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/asset", Description: "删除数据资产", ApiGroup: "大模型平台", Method: "DELETE"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 llmops 菜单（父级目录，子页面菜单随各 web 条目追加）
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
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册 llmops 策略（幂等）：仅 888 全量（平台管理面）
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/llmops/ping", "GET"},
		{"/llmops/asset/list", "GET"},
		{"/llmops/asset", "GET"},
		{"/llmops/asset", "DELETE"},
	} {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			global.GVA_LOG.Error("llmops casbin 添加失败: " + err.Error())
		}
	}
}
