// Package api 白泽 Kubernetes 接口层
package api

import "github.com/hequan2017/new-ops/server/plugin/k8s/service"

var Api = new(api)

type api struct {
	K8sCluster k8sCluster
	K8sResource k8sResource
}

var k8sClusterService = service.Service.K8sClusterService
