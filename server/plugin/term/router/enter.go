// Package router 白泽终端路由
package router

import (
	"github.com/gin-gonic/gin"
)

type RouterGroup struct {
	TerminalRouter
}

var RouterGroupApp = new(RouterGroup)

// Init 注册路由：WS 端点挂 public 组（握手时自验 query token，浏览器 WS 无法带 header）
func (rg *RouterGroup) Init(public, private *gin.RouterGroup) {
	_ = private
	rg.TerminalRouter.InitTerminalRouter(public)
}
