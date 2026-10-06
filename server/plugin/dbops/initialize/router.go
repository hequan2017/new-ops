// Package initialize 白泽数据库工单插件初始化
package initialize

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/global"
	sysModel "github.com/hequan2017/new-ops/server/model/system"
	pluginUtils "github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
	srvUtils "github.com/hequan2017/new-ops/server/utils"
	"github.com/hequan2017/new-ops/server/plugin/dbops/model"
	"github.com/hequan2017/new-ops/server/plugin/dbops/router"
	"github.com/pkg/errors"
)

// Gorm 注册 dbops 数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.DbopsInstance),
		new(model.DbopsOrder),
	)
	if err != nil {
		err = errors.Wrap(err, "dbops 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册 dbops 路由（private 组，casbin 鉴权）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	router.RouterGroupApp.Init(private)
}

// Api 注册 dbops API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/dbops/instance", Description: "创建实例", ApiGroup: "数据库工单", Method: "POST"},
		{Path: "/dbops/instance", Description: "更新实例", ApiGroup: "数据库工单", Method: "PUT"},
		{Path: "/dbops/instance", Description: "删除实例", ApiGroup: "数据库工单", Method: "DELETE"},
		{Path: "/dbops/instance/list", Description: "实例列表", ApiGroup: "数据库工单", Method: "GET"},
		{Path: "/dbops/instance/test", Description: "实例连通检测", ApiGroup: "数据库工单", Method: "POST"},
		{Path: "/dbops/order", Description: "创建 SQL 工单", ApiGroup: "数据库工单", Method: "POST"},
		{Path: "/dbops/order/audit", Description: "SQL 审核", ApiGroup: "数据库工单", Method: "POST"},
		{Path: "/dbops/order/cancel", Description: "取消工单", ApiGroup: "数据库工单", Method: "POST"},
		{Path: "/dbops/order/list", Description: "工单列表", ApiGroup: "数据库工单", Method: "POST"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 dbops 菜单
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "dbops",
			Name:      "dbops",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      9,
			Meta:      sysModel.Meta{Title: "数据库工单", Icon: "coin"},
		},
		{
			Path:      "dbopsInstance",
			Name:      "dbopsInstance",
			Hidden:    false,
			Component: "plugin/dbops/view/instance/index.vue",
			Sort:      1,
			Meta:      sysModel.Meta{Title: "实例管理", Icon: "server"},
		},
		{
			Path:      "dbopsOrder",
			Name:      "dbopsOrder",
			Hidden:    false,
			Component: "plugin/dbops/view/order/index.vue",
			Sort:      2,
			Meta:      sysModel.Meta{Title: "SQL 工单", Icon: "document-add"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册 dbops 策略（幂等）：888 全量；9528 只读（list）
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/dbops/instance", "POST"},
		{"/dbops/instance", "PUT"},
		{"/dbops/instance", "DELETE"},
		{"/dbops/instance/list", "GET"},
		{"/dbops/instance/test", "POST"},
		{"/dbops/order", "POST"},
		{"/dbops/order/audit", "POST"},
		{"/dbops/order/cancel", "POST"},
		{"/dbops/order/list", "POST"},
	} {
		roles := []string{"888"}
		if p.Path == "/dbops/instance/list" || p.Path == "/dbops/order/list" {
			roles = append(roles, "9528")
		}
		for _, role := range roles {
			has, err := e.HasPolicy(role, p.Path, p.Method)
			if err != nil || has {
				continue
			}
			if _, err := e.AddPolicy(role, p.Path, p.Method); err != nil {
				global.GVA_LOG.Error("dbops casbin 添加失败: " + err.Error())
			}
		}
	}
}
