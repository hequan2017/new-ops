// Package router oplib 路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/oplib/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 oplib 路由（private 组，JWT+casbin 由 initialize.Router 显式挂载）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	op := Router.Group("oplib/operator")
	{
		op.GET("list", v1.Api.GetOperators)
		op.POST("", v1.Api.SaveOperator)
		op.DELETE("", v1.Api.DeleteOperator)
	}
}
