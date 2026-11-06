// Package router llmops 路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/llmops/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 llmops 路由（private 组，JWT+casbin 由 initialize.Router 显式挂载）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	base := Router.Group("llmops")
	{
		base.GET("ping", v1.Api.Ping)
	}
	asset := Router.Group("llmops/asset")
	{
		asset.GET("list", v1.Api.GetAssets)
		asset.GET("", v1.Api.GetAsset)
		asset.POST("importDoc", v1.Api.ImportDoc)
		asset.POST("importWeb", v1.Api.ImportWeb)
		asset.POST("importQa", v1.Api.ImportQa)
		asset.POST("importSemi", v1.Api.ImportSemi)
		asset.DELETE("", v1.Api.DeleteAsset)
	}
	ds := Router.Group("llmops/datasource")
	{
		ds.GET("list", v1.Api.GetDataSources)
		ds.POST("", v1.Api.SaveDataSource)
		ds.POST("test", v1.Api.TestDataSource)
		ds.GET("tables", v1.Api.ListDataSourceTables)
		ds.GET("preview", v1.Api.PreviewDataSourceTable)
		ds.DELETE("", v1.Api.DeleteDataSource)
	}
}
