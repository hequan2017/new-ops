// Package router 白泽流水线路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/pipeline/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册流水线路由（private 组，casbin 鉴权）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	pl := Router.Group("pipeline")
	{
		pl.POST("", v1.PipelineApi.CreatePipeline)
		pl.PUT("", v1.PipelineApi.UpdatePipeline)
		pl.DELETE("", v1.PipelineApi.DeletePipeline)
		pl.GET("list", v1.PipelineApi.GetPipelineList)
		pl.GET("find", v1.PipelineApi.GetPipelineDetail)
	}
}
