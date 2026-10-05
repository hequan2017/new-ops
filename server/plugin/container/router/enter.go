// Package router 白泽容器管理路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/container/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册容器管理路由（private 组，casbin 鉴权）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
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
		cc.GET("list", v1.ContainerApi.ListContainers)
		cc.POST("action", v1.ContainerApi.ContainerAction)
	}
}
