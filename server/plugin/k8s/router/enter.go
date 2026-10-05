// Package router 白泽 Kubernetes 路由
package router

import (
	"github.com/gin-gonic/gin"
)

type RouterGroup struct {
	K8sClusterRouter
}

var RouterGroupApp = new(RouterGroup)

// Init 注册 k8s 路由（private 组，casbin 鉴权）
func (rg *RouterGroup) Init(public, private *gin.RouterGroup) {
	_ = public
	rg.K8sClusterRouter.InitK8sClusterRouter(private)
}
