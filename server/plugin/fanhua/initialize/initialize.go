// Package initialize fanhua 繁花-文科大模型插件初始化（表/API/菜单/授权/路由，全部幂等）
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

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/router"
)

// Gorm 注册 fanhua 数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.FanhuaCorpus),
		new(model.FanhuaParallel),
	)
	if err != nil {
		err = errors.Wrap(err, "fanhua 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册 fanhua 路由（private 组显式挂 JWT+casbin）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}

// Api 注册 fanhua API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/fanhua/ping", Description: "繁花文科大模型连通性探测", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/corpus/list", Description: "沪语语料列表", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/corpus", Description: "保存沪语语料", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/corpus/import", Description: "批量导入语料", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/corpus", Description: "删除沪语语料", ApiGroup: "繁花文科大模型", Method: "DELETE"},
		{Path: "/fanhua/parallel/list", Description: "平行语料列表", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/parallel/search", Description: "普通话对照检索", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/parallel", Description: "保存平行语料", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/parallel", Description: "删除平行语料", ApiGroup: "繁花文科大模型", Method: "DELETE"},
		{Path: "/fanhua/clean/dry", Description: "清洗试运行", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/clean", Description: "语料批量清洗", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/tokenize", Description: "沪语分词", ApiGroup: "繁花文科大模型", Method: "POST"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 fanhua 菜单（繁花文科大模型父级，子页面随 N16 挂其下）
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "fanhua",
			Name:      "fanhua",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      7,
			Meta:      sysModel.Meta{Title: "繁花文科大模型", Icon: "notebook-2"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册 fanhua 策略（幂等）：仅 888 全量
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/fanhua/ping", "GET"},
		{"/fanhua/corpus/list", "GET"},
		{"/fanhua/corpus", "POST"},
		{"/fanhua/corpus/import", "POST"},
		{"/fanhua/corpus", "DELETE"},
		{"/fanhua/parallel/list", "GET"},
		{"/fanhua/parallel/search", "GET"},
		{"/fanhua/parallel", "POST"},
		{"/fanhua/parallel", "DELETE"},
		{"/fanhua/clean/dry", "POST"},
		{"/fanhua/clean", "POST"},
		{"/fanhua/tokenize", "POST"},
	} {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			global.GVA_LOG.Error("fanhua casbin 添加失败: " + err.Error())
		}
	}
}
