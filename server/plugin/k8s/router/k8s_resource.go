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
		res.GET("pod/detail", v1.Api.K8sResource.GetPodDetail)
		res.GET("deployment/list", v1.Api.K8sResource.ListDeployments)
		res.GET("statefulset/list", v1.Api.K8sResource.ListStatefulSets)
		res.GET("daemonset/list", v1.Api.K8sResource.ListDaemonSets)
		res.GET("workload/yaml", v1.Api.K8sResource.GetWorkloadYAML)
		res.GET("node/list", v1.Api.K8sResource.ListNodes)
		res.GET("service/list", v1.Api.K8sResource.ListServices)
		res.GET("configmap/list", v1.Api.K8sResource.ListConfigMaps)
		res.GET("secret/list", v1.Api.K8sResource.ListSecrets)
		res.GET("pvc/list", v1.Api.K8sResource.ListPVCs)
		res.GET("ingress/list", v1.Api.K8sResource.ListIngresses)
		res.GET("event/list", v1.Api.K8sResource.ListEvents)
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
	podGroup := Router.Group("k8s/pod")
	{
		podGroup.POST("delete", v1.Api.K8sWrite.DeletePod)
	}
	yamlGroup := Router.Group("k8s/workload")
	{
		yamlGroup.POST("diff", v1.Api.K8sYaml.PreviewWorkloadYAML)
		yamlGroup.POST("apply", v1.Api.K8sYaml.ApplyWorkloadYAML)
	}
}

// K8sHelmRouter Helm release 管理路由
type K8sHelmRouter struct{}

// InitK8sHelmRouter Helm 路由（list/history 只读；install/uninstall/rollback 写级 888）
func (r *K8sHelmRouter) InitK8sHelmRouter(Router *gin.RouterGroup) {
	helm := Router.Group("k8s/helm")
	{
		helm.GET("list", v1.Api.K8sHelm.ListHelmReleases)
		helm.GET("history", v1.Api.K8sHelm.GetHelmHistory)
		helm.POST("install", v1.Api.K8sHelm.InstallHelmRelease)
		helm.POST("uninstall", v1.Api.K8sHelm.UninstallHelmRelease)
		helm.POST("rollback", v1.Api.K8sHelm.RollbackHelmRelease)
	}
}
