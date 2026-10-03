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
}
