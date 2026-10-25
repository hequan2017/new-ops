// Package service k8s 资源浏览辅助（client-go 封装）
package service

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
)

// clientcmdRESTConfig 从 kubeconfig 文本构造 RESTConfig
func clientcmdRESTConfig(kubeconfig string) (*rest.Config, error) {
	return clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
}

// kubernetesNewForConfig 从 RESTConfig 构造 clientset
func kubernetesNewForRESTConfig(cfg *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(cfg)
}

// statusReady 转 int32（零值兼容）
func statusReady(v int32) int32 { return v }

// derefInt32 安全解引用
func derefInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}

// internalIP 取节点 InternalIP
func internalIP(addrs []corev1.NodeAddress) string {
	for _, a := range addrs {
		if a.Type == corev1.NodeInternalIP {
			return a.Address
		}
	}
	for _, a := range addrs {
		if a.Type == corev1.NodeHostName {
			return a.Address
		}
	}
	return "-"
}
