// Package initialize ziqiang 自强-理工科大模型插件初始化（表/API/菜单/授权/路由，全部幂等）
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

	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/router"
)

// Gorm 注册 ziqiang 数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.ZqTextCorpus),
		new(model.ZqComparePair),
	)
	if err != nil {
		err = errors.Wrap(err, "ziqiang 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
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
		{Path: "/ziqiang/textCorpus/list", Description: "文本语料列表", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/textCorpus/import", Description: "批量导入文本语料", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/textCorpus", Description: "保存文本语料", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/textCorpus", Description: "删除文本语料", ApiGroup: "自强理工科大模型", Method: "DELETE"},
		{Path: "/ziqiang/comparePair/list", Description: "对比样本对列表", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/comparePair", Description: "保存对比样本对", ApiGroup: "自强理工科大模型", Method: "POST"},
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
		{"/ziqiang/textCorpus/list", "GET"},
		{"/ziqiang/textCorpus/import", "POST"},
		{"/ziqiang/textCorpus", "POST"},
		{"/ziqiang/textCorpus", "DELETE"},
		{"/ziqiang/comparePair/list", "GET"},
		{"/ziqiang/comparePair", "POST"},
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
