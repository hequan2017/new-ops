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
		new(model.LlmopsQaPair),
		new(model.LlmopsDataSource),
		new(model.LlmopsCategory),
		new(model.LlmopsAssetEvent),
		new(model.LlmopsModel),
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
		{Path: "/llmops/asset/importDoc", Description: "文档采集上传", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/asset/importWeb", Description: "网页采集", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/asset/importQa", Description: "问答对采集", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/asset/importSemi", Description: "半结构化编目", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/asset/categorize", Description: "资产挂分类/打标签", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/asset/events", Description: "资产事件流水", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/asset/importMedia", Description: "媒体元数据编目", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/asset/transcribe", Description: "媒体 ASR 转写", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/category", Description: "保存编目分类", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/category/tree", Description: "编目分类树", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/category", Description: "删除编目分类", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/model/list", Description: "模型注册列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/model", Description: "保存模型注册", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/model", Description: "删除模型注册", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/datasource/list", Description: "数据源列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/datasource", Description: "保存数据源", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/datasource/test", Description: "数据源连接测试", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/datasource/tables", Description: "库表清单", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/datasource/preview", Description: "表采样预览", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/datasource", Description: "删除数据源", ApiGroup: "大模型平台", Method: "DELETE"},
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
		{
			Path:      "llmopsData",
			Name:      "llmopsData",
			Hidden:    false,
			Component: "plugin/llmops/view/data/index.vue",
			Sort:      1,
			Meta:      sysModel.Meta{Title: "数据管理", Icon: "files"},
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
		{"/llmops/asset/importDoc", "POST"},
		{"/llmops/asset/importWeb", "POST"},
		{"/llmops/asset/importQa", "POST"},
		{"/llmops/asset/importSemi", "POST"},
		{"/llmops/asset/categorize", "POST"},
		{"/llmops/asset/events", "GET"},
		{"/llmops/asset/importMedia", "POST"},
		{"/llmops/asset/transcribe", "POST"},
		{"/llmops/category", "POST"},
		{"/llmops/category/tree", "GET"},
		{"/llmops/category", "DELETE"},
		{"/llmops/model/list", "GET"},
		{"/llmops/model", "POST"},
		{"/llmops/model", "DELETE"},
		{"/llmops/datasource/list", "GET"},
		{"/llmops/datasource", "POST"},
		{"/llmops/datasource/test", "POST"},
		{"/llmops/datasource/tables", "GET"},
		{"/llmops/datasource/preview", "GET"},
		{"/llmops/datasource", "DELETE"},
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
