// Package router 白泽容器管理路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/container/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册容器管理路由：操作走 private（casbin）；WS 流挂 public（query token 自验，与 term 一致）
func (rg *RouterGroup) Init(Router, Public *gin.RouterGroup) {
	ep := Router.Group("container/endpoint")
	{
		ep.POST("", v1.ContainerApi.CreateEndpoint)
		ep.PUT("", v1.ContainerApi.UpdateEndpoint)
		ep.DELETE("", v1.ContainerApi.DeleteEndpoint)
		ep.GET("list", v1.ContainerApi.GetEndpointList)
		ep.POST("check", v1.ContainerApi.CheckEndpoint)
	}
	cc := Router.Group("container/container")
	{
		cc.POST("", v1.ContainerApi.CreateContainer)
		cc.GET("list", v1.ContainerApi.ListContainers)
		cc.POST("action", v1.ContainerApi.ContainerAction)
	}
	Router.GET("container/event/list", v1.ContainerApi.GetEventList)
	img := Router.Group("container/image")
	{
		img.GET("list", v1.ContainerApi.ListImages)
		img.POST("pull", v1.ContainerApi.PullImage)
		img.GET("pull-status", v1.ContainerApi.PullStatus)
		img.DELETE("", v1.ContainerApi.RemoveImage)
	}
	nw := Router.Group("container/network")
	{
		nw.GET("list", v1.ContainerApi.ListNetworks)
		nw.POST("", v1.ContainerApi.CreateNetwork)
		nw.DELETE("", v1.ContainerApi.RemoveNetwork)
	}
	vol := Router.Group("container/volume")
	{
		vol.GET("list", v1.ContainerApi.ListVolumes)
		vol.DELETE("", v1.ContainerApi.RemoveVolume)
	}
	Router.GET("container/container/stats", v1.ContainerApi.ContainerStats)
	Public.GET("container/container/logws", v1.ContainerApi.LogsWS)
	Public.GET("container/container/execws", v1.ContainerApi.ExecWS)
}
