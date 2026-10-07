// Package initialize 白泽 GPU 算力插件初始化
package initialize

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/global"
	sysModel "github.com/hequan2017/new-ops/server/model/system"
	srvUtils "github.com/hequan2017/new-ops/server/utils"
	pluginUtils "github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
	"github.com/hequan2017/new-ops/server/plugin/gpu/model"
	"github.com/hequan2017/new-ops/server/plugin/gpu/router"
	"github.com/pkg/errors"
)

// Gorm 注册 GPU 数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.GpuNode),
		new(model.GpuSpec),
		new(model.GpuInstance),
	)
	if err != nil {
		err = errors.Wrap(err, "gpu 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册 GPU 路由（private 组，casbin 鉴权）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	router.RouterGroupApp.Init(private)
}

// Api 注册 GPU API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/gpu/node", Description: "注册算力节点", ApiGroup: "GPU 算力", Method: "POST"},
		{Path: "/gpu/node", Description: "更新算力节点", ApiGroup: "GPU 算力", Method: "PUT"},
		{Path: "/gpu/node", Description: "删除算力节点", ApiGroup: "GPU 算力", Method: "DELETE"},
		{Path: "/gpu/node/list", Description: "节点列表（含余量）", ApiGroup: "GPU 算力", Method: "GET"},
		{Path: "/gpu/spec", Description: "创建规格", ApiGroup: "GPU 算力", Method: "POST"},
		{Path: "/gpu/spec", Description: "删除规格", ApiGroup: "GPU 算力", Method: "DELETE"},
		{Path: "/gpu/spec/list", Description: "规格列表", ApiGroup: "GPU 算力", Method: "GET"},
		{Path: "/gpu/instance", Description: "开通 GPU 实例", ApiGroup: "GPU 算力", Method: "POST"},
		{Path: "/gpu/instance/release", Description: "销毁 GPU 实例", ApiGroup: "GPU 算力", Method: "POST"},
		{Path: "/gpu/instance/list", Description: "实例列表", ApiGroup: "GPU 算力", Method: "GET"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 GPU 菜单
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "gpu",
			Name:      "gpu",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      10,
			Meta:      sysModel.Meta{Title: "GPU 算力", Icon: "video-card"},
		},
		{
			Path:      "gpuInstance",
			Name:      "gpuInstance",
			Hidden:    false,
			Component: "plugin/gpu/view/instance/index.vue",
			Sort:      1,
			Meta:      sysModel.Meta{Title: "算力管理", Icon: "box"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册 GPU 策略（幂等）：888 全量；9528 只读（list）
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/gpu/node", "POST"},
		{"/gpu/node", "PUT"},
		{"/gpu/node", "DELETE"},
		{"/gpu/node/list", "GET"},
		{"/gpu/spec", "POST"},
		{"/gpu/spec", "DELETE"},
		{"/gpu/spec/list", "GET"},
		{"/gpu/instance", "POST"},
		{"/gpu/instance/release", "POST"},
		{"/gpu/instance/list", "GET"},
	} {
		roles := []string{"888"}
		if p.Method == "GET" {
			roles = append(roles, "9528")
		}
		for _, role := range roles {
			has, err := e.HasPolicy(role, p.Path, p.Method)
			if err != nil || has {
				continue
			}
			if _, err := e.AddPolicy(role, p.Path, p.Method); err != nil {
				global.GVA_LOG.Error("gpu casbin 添加失败: " + err.Error())
			}
		}
	}
}
