// Package initialize aiops AI 运维插件初始化（表/API/菜单/授权/路由）
package initialize

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/middleware"
	pluginUtils "github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
	srvUtils "github.com/hequan2017/new-ops/server/utils"
	sysModel "github.com/hequan2017/new-ops/server/model/system"

	"github.com/hequan2017/new-ops/server/plugin/aiops/model"
	"github.com/hequan2017/new-ops/server/plugin/aiops/router"
)

// Gorm 注册 aiops 数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.AiopsLlmProvider),
		new(model.AiopsDiagnosis),
	)
	if err != nil {
		err = errors.Wrap(err, "aiops 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册 aiops 路由（private 组显式挂 JWT+casbin）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}

// Api 注册 aiops API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/aiops/provider/list", Description: "LLM 供应商列表", ApiGroup: "AI 运维", Method: "GET"},
		{Path: "/aiops/provider", Description: "保存 LLM 供应商", ApiGroup: "AI 运维", Method: "POST"},
		{Path: "/aiops/provider", Description: "删除 LLM 供应商", ApiGroup: "AI 运维", Method: "DELETE"},
		{Path: "/aiops/llm/chat", Description: "LLM 对话调用", ApiGroup: "AI 运维", Method: "POST"},
		{Path: "/aiops/diagnose", Description: "Pod AI 诊断", ApiGroup: "AI 运维", Method: "POST"},
		{Path: "/aiops/diagnosis/list", Description: "诊断历史列表", ApiGroup: "AI 运维", Method: "POST"},
		{Path: "/aiops/diagnosis", Description: "诊断详情", ApiGroup: "AI 运维", Method: "GET"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 aiops 菜单
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "aiops",
			Name:      "aiops",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      10,
			Meta:      sysModel.Meta{Title: "AI 运维", Icon: "magic-stick"},
		},
		{
			Path:      "aiopsDiagnose",
			Name:      "aiopsDiagnose",
			Hidden:    false,
			Component: "plugin/aiops/view/diagnose/index.vue",
			Sort:      1,
			Meta:      sysModel.Meta{Title: "Pod 诊断", Icon: "first-aid-kit"},
		},
		{
			Path:      "aiopsLlm",
			Name:      "aiopsLlm",
			Hidden:    false,
			Component: "plugin/aiops/view/llm/index.vue",
			Sort:      2,
			Meta:      sysModel.Meta{Title: "模型网关", Icon: "cpu"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册 aiops 策略（幂等）：仅 888 全量（AI 诊断与密钥管理均为管理员面）
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/aiops/provider/list", "GET"},
		{"/aiops/provider", "POST"},
		{"/aiops/provider", "DELETE"},
		{"/aiops/llm/chat", "POST"},
		{"/aiops/diagnose", "POST"},
		{"/aiops/diagnosis/list", "POST"},
		{"/aiops/diagnosis", "GET"},
	} {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			global.GVA_LOG.Error("aiops casbin 添加失败: " + err.Error())
		}
	}
}
