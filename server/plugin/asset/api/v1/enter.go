// Package api asset 插件接口层
package api

import "github.com/hequan2017/new-ops/server/plugin/asset/service"

var Api = new(api)

type api struct {
	AssetHost assetHost
}

var assetHostService = service.Service.AssetHostService
