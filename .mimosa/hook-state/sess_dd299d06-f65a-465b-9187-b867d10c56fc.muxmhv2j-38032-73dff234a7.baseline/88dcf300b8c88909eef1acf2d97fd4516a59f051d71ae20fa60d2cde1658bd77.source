// Package router 白泽 GPU 算力路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/gpu/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 GPU 路由（private 组，casbin 鉴权）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	node := Router.Group("gpu/node")
	{
		node.POST("", v1.GpuApi.CreateNode)
		node.PUT("", v1.GpuApi.UpdateNode)
		node.DELETE("", v1.GpuApi.DeleteNode)
		node.GET("list", v1.GpuApi.ListNode)
	}
	spec := Router.Group("gpu/spec")
	{
		spec.POST("", v1.GpuApi.CreateSpec)
		spec.DELETE("", v1.GpuApi.DeleteSpec)
		spec.GET("list", v1.GpuApi.ListSpec)
	}
	inst := Router.Group("gpu/instance")
	{
		inst.POST("", v1.GpuApi.StartInstance)
		inst.POST("release", v1.GpuApi.ReleaseInstance)
		inst.GET("list", v1.GpuApi.ListInstance)
	}
}
