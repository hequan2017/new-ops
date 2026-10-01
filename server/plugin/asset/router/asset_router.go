// Package router 机房/机柜/产品线路由
package router

import (
	"github.com/gin-gonic/gin"

	assetApi "github.com/hequan2017/new-ops/server/plugin/asset/api/v1"
)

type AssetRoomRouter struct{}
type AssetRackRouter struct{}
type AssetProductLineRouter struct{}

// InitAssetRoomRouter 机房路由
func (r *AssetRoomRouter) InitAssetRoomRouter(Router *gin.RouterGroup) {
	roomRouter := Router.Group("asset/room")
	{
		roomRouter.POST("create", assetApi.Api.AssetRoom.CreateAssetRoom)
		roomRouter.DELETE("delete", assetApi.Api.AssetRoom.DeleteAssetRoom)
		roomRouter.PUT("update", assetApi.Api.AssetRoom.UpdateAssetRoom)
		roomRouter.GET("list", assetApi.Api.AssetRoom.GetAssetRoomList)
	}
}

// InitAssetRackRouter 机柜路由
func (r *AssetRackRouter) InitAssetRackRouter(Router *gin.RouterGroup) {
	rackRouter := Router.Group("asset/rack")
	{
		rackRouter.POST("create", assetApi.Api.AssetRack.CreateAssetRack)
		rackRouter.DELETE("delete", assetApi.Api.AssetRack.DeleteAssetRack)
		rackRouter.PUT("update", assetApi.Api.AssetRack.UpdateAssetRack)
		rackRouter.GET("list", assetApi.Api.AssetRack.GetAssetRackList)
	}
}

// InitAssetProductLineRouter 产品线路由
func (r *AssetProductLineRouter) InitAssetProductLineRouter(Router *gin.RouterGroup) {
	plRouter := Router.Group("asset/productLine")
	{
		plRouter.POST("create", assetApi.Api.AssetProductLine.CreateAssetProductLine)
		plRouter.DELETE("delete", assetApi.Api.AssetProductLine.DeleteAssetProductLine)
		plRouter.PUT("update", assetApi.Api.AssetProductLine.UpdateAssetProductLine)
		plRouter.GET("list", assetApi.Api.AssetProductLine.GetAssetProductLineList)
	}
}
