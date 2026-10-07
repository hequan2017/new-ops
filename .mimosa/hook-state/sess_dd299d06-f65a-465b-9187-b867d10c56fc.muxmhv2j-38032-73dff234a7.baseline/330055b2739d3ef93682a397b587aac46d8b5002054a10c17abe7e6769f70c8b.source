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

// k8sPodLogOptions corev1.PodLogOptions 轻量构造
type k8sPodLogOptions struct {
	Container string
	TailLines int64
}

func (o *k8sPodLogOptions) podLogOptions() *corev1.PodLogOptions {
	opts := &corev1.PodLogOptions{}
	if o.Container != "" {
		opts.Container = o.Container
	}
	if o.TailLines > 0 {
		t := o.TailLines
		opts.TailLines = &t
	}
	return opts
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
