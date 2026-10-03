// Package router 白泽终端路由
package router

import (
	"github.com/gin-gonic/gin"
)

type RouterGroup struct {
	TerminalRouter
	TermSessionRouter
	SftpRouter
}

var RouterGroupApp = new(RouterGroup)

// Init 注册路由：WS 端点挂 public 组（握手时自验 query token），审计查询挂 private+casbin
func (rg *RouterGroup) Init(public, private *gin.RouterGroup) {
	rg.TerminalRouter.InitTerminalRouter(public)
	rg.TermSessionRouter.InitTermSessionRouter(private)
	rg.SftpRouter.InitSftpRouter(private)
}
