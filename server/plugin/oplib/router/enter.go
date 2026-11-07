// Package router oplib 路由
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/middleware"

	v1 "github.com/hequan2017/new-ops/server/plugin/oplib/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 oplib 路由（private 组，JWT+casbin 由 initialize.Router 显式挂载）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	pl := Router.Group("oplib/pipeline")
	{
		pl.GET("list", v1.Api.GetPipelines)
		pl.POST("", middleware.OperationRecord(), v1.Api.SavePipeline)
		pl.POST("run", middleware.OperationRecord(), v1.Api.RunPipeline)
		pl.DELETE("", middleware.OperationRecord(), v1.Api.DeletePipeline)
	}
	op := Router.Group("oplib/operator")
	{
		op.GET("list", v1.Api.GetOperators)
		op.POST("", middleware.OperationRecord(), v1.Api.SaveOperator)
		op.POST("run", middleware.OperationRecord(), v1.Api.RunOperator)
		op.GET("runLogs", v1.Api.GetRunLogs)
		op.DELETE("", middleware.OperationRecord(), v1.Api.DeleteOperator)
	}
}
