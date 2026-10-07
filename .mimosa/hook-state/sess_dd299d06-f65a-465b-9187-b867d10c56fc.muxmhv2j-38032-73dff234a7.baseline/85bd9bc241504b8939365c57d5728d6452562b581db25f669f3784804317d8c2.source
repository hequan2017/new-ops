// Package initialize 白泽工单引擎插件初始化
package initialize

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/global"
	sysModel "github.com/hequan2017/new-ops/server/model/system"
	pluginUtils "github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
	srvUtils "github.com/hequan2017/new-ops/server/utils"
	"github.com/hequan2017/new-ops/server/plugin/workflow/model"
	"github.com/hequan2017/new-ops/server/plugin/workflow/router"
	"github.com/pkg/errors"
)

// Gorm 注册工单数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.WfDefinition),
		new(model.WfInstance),
		new(model.WfActionLog),
	)
	if err != nil {
		err = errors.Wrap(err, "workflow 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册工单路由（private 组，casbin 鉴权）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	router.RouterGroupApp.Init(private)
}

// Api 注册工单 API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/workflow/definition", Description: "创建工单定义", ApiGroup: "工单引擎", Method: "POST"},
		{Path: "/workflow/definition", Description: "更新工单定义", ApiGroup: "工单引擎", Method: "PUT"},
		{Path: "/workflow/definition", Description: "删除工单定义", ApiGroup: "工单引擎", Method: "DELETE"},
		{Path: "/workflow/definition/list", Description: "工单定义列表", ApiGroup: "工单引擎", Method: "GET"},
		{Path: "/workflow/instance", Description: "发起工单", ApiGroup: "工单引擎", Method: "POST"},
		{Path: "/workflow/instance/action", Description: "工单流转动作", ApiGroup: "工单引擎", Method: "POST"},
		{Path: "/workflow/instance/list", Description: "工单实例列表", ApiGroup: "工单引擎", Method: "POST"},
		{Path: "/workflow/instance/logs", Description: "工单流转记录", ApiGroup: "工单引擎", Method: "GET"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册工单菜单
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "workflow",
			Name:      "workflow",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      8,
			Meta:      sysModel.Meta{Title: "工单中心", Icon: "tickets"},
		},
		{
			Path:      "workflowTicket",
			Name:      "workflowTicket",
			Hidden:    false,
			Component: "plugin/workflow/view/ticket/index.vue",
			Sort:      1,
			Meta:      sysModel.Meta{Title: "我的工单", Icon: "document-checked"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册工单策略（幂等）：888 全量；9528 只读（列表/记录）
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/workflow/definition", "POST"},
		{"/workflow/definition", "PUT"},
		{"/workflow/definition", "DELETE"},
		{"/workflow/definition/list", "GET"},
		{"/workflow/instance", "POST"},
		{"/workflow/instance/action", "POST"},
		{"/workflow/instance/list", "POST"},
		{"/workflow/instance/logs", "GET"},
	} {
		roles := []string{"888"}
		if p.Method == "GET" || p.Path == "/workflow/instance/list" {
			roles = append(roles, "9528")
		}
		for _, role := range roles {
			has, err := e.HasPolicy(role, p.Path, p.Method)
			if err != nil || has {
				continue
			}
			if _, err := e.AddPolicy(role, p.Path, p.Method); err != nil {
				global.GVA_LOG.Error("workflow casbin 添加失败: " + err.Error())
			}
		}
	}
}
