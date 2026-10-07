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
		res.GET("service/list", v1.Api.K8sResource.ListServices)
		res.GET("configmap/list", v1.Api.K8sResource.ListConfigMaps)
		res.GET("secret/list", v1.Api.K8sResource.ListSecrets)
	}
}

type K8sWriteRouter struct{}

// K8sNodeRouter Node 管理与集群总览路由
type K8sNodeRouter struct{}

// InitK8sNodeRouter Node 管理与集群总览路由
func (r *K8sNodeRouter) InitK8sNodeRouter(Router *gin.RouterGroup) {
	node := Router.Group("k8s")
	{
		node.GET("cluster/overview", v1.Api.K8sNode.GetClusterOverview)
		node.GET("node/detail", v1.Api.K8sNode.GetNodeDetail)
	}
	writeGroup := Router.Group("k8s/node")
	{
		writeGroup.POST("cordon", v1.Api.K8sNode.CordonNode)
		writeGroup.POST("drain", v1.Api.K8sNode.DrainNode)
	}
}

// InitK8sWriteRouter 写操作路由（private 组；casbin 策略仅 888）
func (r *K8sWriteRouter) InitK8sWriteRouter(Router *gin.RouterGroup) {
	writeGroup := Router.Group("k8s/deployment")
	{
		writeGroup.POST("scale", v1.Api.K8sWrite.ScaleDeployment)
		writeGroup.POST("restart", v1.Api.K8sWrite.RestartDeployment)
	}
}
