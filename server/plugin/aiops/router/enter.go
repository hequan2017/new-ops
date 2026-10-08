// Package router aiops 路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/aiops/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 aiops 路由（private 组，casbin 鉴权）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	provider := Router.Group("aiops/provider")
	{
		provider.GET("list", v1.Api.GetProviders)
		provider.POST("", v1.Api.SaveProvider)
		provider.DELETE("", v1.Api.DeleteProvider)
	}
	llm := Router.Group("aiops/llm")
	{
		llm.POST("chat", v1.Api.Chat)
	}
	diag := Router.Group("aiops")
	{
		diag.POST("diagnose", v1.Api.Diagnose)
		diag.POST("diagnosis/list", v1.Api.GetDiagnosisList)
		diag.GET("diagnosis", v1.Api.GetDiagnosis)
	}
}
