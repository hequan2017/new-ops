// Package router crossauth 路由
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/middleware"

	v1 "github.com/hequan2017/new-ops/server/plugin/crossauth/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 crossauth 路由（private 组，JWT+casbin 由 initialize.Router 显式挂载）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	base := Router.Group("crossauth")
	{
		base.GET("ping", v1.Api.Ping)
	}
	fed := Router.Group("crossauth/federation")
	{
		fed.GET("list", v1.Api.GetCaFederations)
		fed.GET("tasks", v1.Api.GetCaSyncTasks)
		fed.POST("", middleware.OperationRecord(), v1.Api.SaveCaFederation)
		fed.POST("sync", middleware.OperationRecord(), v1.Api.RunCaSyncTask)
	}
	Router.Group("crossauth/syncUser").GET("list", v1.Api.GetCaSyncUsers)
}
