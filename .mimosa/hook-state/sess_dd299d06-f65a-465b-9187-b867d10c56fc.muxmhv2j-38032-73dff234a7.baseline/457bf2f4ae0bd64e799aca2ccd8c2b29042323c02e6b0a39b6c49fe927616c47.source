// Package router 白泽工单引擎路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/workflow/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册工单路由（private 组，casbin 鉴权）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	def := Router.Group("workflow/definition")
	{
		def.POST("", v1.WorkflowApi.CreateDefinition)
		def.PUT("", v1.WorkflowApi.UpdateDefinition)
		def.DELETE("", v1.WorkflowApi.DeleteDefinition)
		def.GET("list", v1.WorkflowApi.ListDefinitions)
	}
	ins := Router.Group("workflow/instance")
	{
		ins.POST("", v1.WorkflowApi.StartInstance)
		ins.POST("action", v1.WorkflowApi.SubmitAction)
		ins.POST("list", v1.WorkflowApi.ListInstances)
		ins.GET("logs", v1.WorkflowApi.GetInstanceLogs)
	}
}
