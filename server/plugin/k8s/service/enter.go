// Package service 白泽 Kubernetes 多集群服务
package service

// ServiceGroup k8s 插件服务组
type ServiceGroup struct {
	K8sClusterService
}

var Service = new(ServiceGroup)
