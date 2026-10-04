// Package router 终端路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/term/api/v1"
)

type TerminalRouter struct{}

// InitTerminalRouter 注册 WebSSH WebSocket 端点
func (r *TerminalRouter) InitTerminalRouter(Router *gin.RouterGroup) {
	Router.GET("term/ws", v1.Api.Terminal.WebSSH)
	Router.GET("term/logtail", v1.Api.Terminal.LogTail)
}

type TermSessionRouter struct{}

// InitTermSessionRouter 会话审计查询路由（private + casbin 888）
func (r *TermSessionRouter) InitTermSessionRouter(Router *gin.RouterGroup) {
	sessRouter := Router.Group("term/session")
	{
		sessRouter.POST("list", v1.Api.TermSession.GetTermSessionList)
		sessRouter.GET("streams", v1.Api.TermSession.GetTermSessionStreams)
		sessRouter.GET("commands", v1.Api.TermSession.GetTermSessionCommands)
	}
}
