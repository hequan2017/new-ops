// Package router 白泽 Kubernetes 路由
package router

import (
	"github.com/gin-gonic/gin"
)

type RouterGroup struct {
	K8sClusterRouter
	K8sResourceRouter
	K8sWriteRouter
}

var RouterGroupApp = new(RouterGroup)

// Init 注册 k8s 路由（private 组，casbin 鉴权）
func (rg *RouterGroup) Init(public, private *gin.RouterGroup) {
	_ = public
	rg.K8sClusterRouter.InitK8sClusterRouter(private)
	rg.K8sResourceRouter.InitK8sResourceRouter(private)
	rg.K8sWriteRouter.InitK8sWriteRouter(private)
}
