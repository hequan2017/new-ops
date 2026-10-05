// Package router 白泽监控路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/monitor/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册监控路由（private 组，casbin 鉴权）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	m := Router.Group("monitor/metric")
	{
		m.GET("list", v1.MonitorApi.GetMetrics)
		m.POST("collect", v1.MonitorApi.CollectNow)
	}
}
