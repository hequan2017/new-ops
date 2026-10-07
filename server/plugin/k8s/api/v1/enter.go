// Package api 白泽 Kubernetes 接口层
package api

import "github.com/hequan2017/new-ops/server/plugin/k8s/service"

var Api = new(api)

type api struct {
	K8sCluster  k8sCluster
	K8sResource k8sResource
	K8sConfig   k8sConfig
	K8sWrite    k8sWrite
	K8sNode     k8sNode
	K8sYaml     k8sYaml
	K8sHelm     k8sHelm
	K8sGrant    k8sGrant
}

var k8sClusterService = service.Service.K8sClusterService
