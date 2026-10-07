// Package router asset 插件路由
package router

import (
	"github.com/gin-gonic/gin"
)

type RouterGroup struct {
	AssetHostRouter
	AssetRoomRouter
	AssetRackRouter
	AssetProductLineRouter
	AssetGroupRouter
	CredCredentialRouter
}

var RouterGroupApp = new(RouterGroup)

// Init 注册路由：public 组预留（后续探测类接口），private 组走 JWT + Casbin
func (rg *RouterGroup) Init(public, private *gin.RouterGroup) {
	_ = public
	rg.AssetHostRouter.InitAssetHostRouter(private)
	rg.AssetRoomRouter.InitAssetRoomRouter(private)
	rg.AssetRackRouter.InitAssetRackRouter(private)
	rg.AssetProductLineRouter.InitAssetProductLineRouter(private)
	rg.AssetGroupRouter.InitAssetGroupRouter(private)
	rg.CredCredentialRouter.InitCredCredentialRouter(private)
}
