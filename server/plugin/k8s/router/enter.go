// Package router 白泽 Kubernetes 路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/k8s/api/v1"
)

type RouterGroup struct {
	K8sClusterRouter
	K8sResourceRouter
	K8sWriteRouter
	K8sNodeRouter
	K8sHelmRouter
}

var RouterGroupApp = new(RouterGroup)

// Init 注册 k8s 路由（private 组 JWT+casbin；WS 流挂 public 组自验 query token）
func (rg *RouterGroup) Init(public, private *gin.RouterGroup) {
	rg.K8sClusterRouter.InitK8sClusterRouter(private)
	rg.K8sResourceRouter.InitK8sResourceRouter(private)
	rg.K8sWriteRouter.InitK8sWriteRouter(private)
	rg.K8sNodeRouter.InitK8sNodeRouter(private)
	rg.K8sHelmRouter.InitK8sHelmRouter(private)
	public.GET("k8s/pod/execws", v1.Api.K8sResource.PodExecWS)
}
