// Package router 白泽数据库工单路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/dbops/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 dbops 路由（private 组，casbin 鉴权）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	inst := Router.Group("dbops/instance")
	{
		inst.POST("", v1.DbopsApi.CreateInstance)
		inst.PUT("", v1.DbopsApi.UpdateInstance)
		inst.DELETE("", v1.DbopsApi.DeleteInstance)
		inst.GET("list", v1.DbopsApi.ListInstances)
		inst.POST("test", v1.DbopsApi.TestInstance)
	}
	order := Router.Group("dbops/order")
	{
		order.POST("", v1.DbopsApi.CreateOrder)
		order.POST("audit", v1.DbopsApi.AuditOrder)
		order.POST("cancel", v1.DbopsApi.CancelOrder)
		order.POST("list", v1.DbopsApi.ListOrders)
	}
}
