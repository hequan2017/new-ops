// Package api asset 插件接口层
package api

import "github.com/hequan2017/new-ops/server/plugin/asset/service"

var Api = new(api)

type api struct {
	AssetHost      assetHost
	AssetRoom      assetRoom
	AssetRack      assetRack
	AssetProductLine assetProductLine
}

var (
	assetHostService      = service.Service.AssetHostService
	assetRoomService      = service.Service.AssetRoomService
	assetRackService      = service.Service.AssetRackService
	assetProductLineService = service.Service.AssetProductLineService
)
