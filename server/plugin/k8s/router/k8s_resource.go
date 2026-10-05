// Package router 资源浏览路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/k8s/api/v1"
)

type K8sResourceRouter struct{}

// InitK8sResourceRouter 资源浏览路由（只读）
func (r *K8sResourceRouter) InitK8sResourceRouter(Router *gin.RouterGroup) {
	res := Router.Group("k8s")
	{
		res.GET("pod/list", v1.Api.K8sResource.ListPods)
		res.GET("pod/logs", v1.Api.K8sResource.GetPodLogs)
		res.GET("deployment/list", v1.Api.K8sResource.ListDeployments)
		res.GET("node/list", v1.Api.K8sResource.ListNodes)
	}
}
