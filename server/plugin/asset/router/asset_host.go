// Package router 主机资产路由（/asset/host 前缀，错误码段见 DEV_PLAN 3.6）
package router

import (
	"github.com/gin-gonic/gin"

	assetApi "github.com/hequan2017/new-ops/server/plugin/asset/api/v1"
)

type AssetHostRouter struct{}

// InitAssetHostRouter 初始化主机资产路由
func (r *AssetHostRouter) InitAssetHostRouter(Router *gin.RouterGroup) {
	hostRouter := Router.Group("asset/host")
	{
		hostRouter.POST("create", assetApi.Api.AssetHost.CreateAssetHost)
		hostRouter.DELETE("delete", assetApi.Api.AssetHost.DeleteAssetHost)
		hostRouter.DELETE("deleteByIds", assetApi.Api.AssetHost.DeleteAssetHostByIds)
		hostRouter.PUT("update", assetApi.Api.AssetHost.UpdateAssetHost)
		hostRouter.GET("find", assetApi.Api.AssetHost.FindAssetHost)
		hostRouter.GET("history", assetApi.Api.AssetHost.GetAssetHostHistory)
		hostRouter.GET("export", assetApi.Api.AssetHost.ExportAssetHost)
		hostRouter.POST("import", assetApi.Api.AssetHost.ImportAssetHost)
		hostRouter.POST("collect", assetApi.Api.AssetHost.CollectAssetHost)
		hostRouter.POST("list", assetApi.Api.AssetHost.GetAssetHostList)
	}
}
