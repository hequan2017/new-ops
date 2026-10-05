// Package router 集群路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/k8s/api/v1"
)

type K8sClusterRouter struct{}

// InitK8sClusterRouter 集群路由
func (r *K8sClusterRouter) InitK8sClusterRouter(Router *gin.RouterGroup) {
	cluster := Router.Group("k8s/cluster")
	{
		cluster.POST("create", v1.Api.K8sCluster.CreateCluster)
		cluster.DELETE("delete", v1.Api.K8sCluster.DeleteCluster)
		cluster.GET("list", v1.Api.K8sCluster.GetClusterList)
		cluster.GET("test", v1.Api.K8sCluster.TestCluster)
	}
}
