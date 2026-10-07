// Package service K1 集群总览（版本/节点/命名空间/Pod/metrics-server 资源用量）
package service

import (
	"context"
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/k8s/model"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// metricsClientsetForCluster 构造 metrics 客户端（仅构造不发请求；集群未装 metrics-server 时 List 阶段才会失败）
func metricsClientsetForCluster(cfg *rest.Config) (*metricsv.Clientset, error) {
	return metricsv.NewForConfig(cfg)
}

// restConfigForCluster 解密 kubeconfig 构造 RESTConfig（总览需要同时构造 core 与 metrics 客户端）
func restConfigForCluster(clusterID uint) (*rest.Config, model.K8sCluster, error) {
	var c model.K8sCluster
	if err := global.GVA_DB.First(&c, clusterID).Error; err != nil {
		return nil, c, fmt.Errorf("集群不存在: %w", err)
	}
	kubeconfig, err := crypto.Decrypt(c.KubeconfigCipher)
	if err != nil {
		return nil, c, fmt.Errorf("kubeconfig 解密失败: %w", err)
	}
	cfg, err := clientcmdRESTConfig(kubeconfig)
	if err != nil {
		return nil, c, err
	}
	return cfg, c, nil
}

// NodeSummary 节点概要统计
type NodeSummary struct {
	Total    int `json:"total"`
	Ready    int `json:"ready"`
	Cordoned int `json:"cordoned"`
}

// PodSummary Pod 概要统计
type PodSummary struct {
	Total   int `json:"total"`
	Running int `json:"running"`
}

// ResourceUsage 全集群资源用量（metrics-server 汇总，相对可分配量）
type ResourceUsage struct {
	CPUUsedCores     float64 `json:"cpuUsedCores"`
	CPUAllocCores    float64 `json:"cpuAllocCores"`
	CPUPercent       float64 `json:"cpuPercent"`
	MemUsedGB        float64 `json:"memUsedGB"`
	MemAllocGB       float64 `json:"memAllocGB"`
	MemPercent       float64 `json:"memPercent"`
	MetricsAvailable bool    `json:"metricsAvailable"`
}

// ClusterOverview 集群总览
type ClusterOverview struct {
	Version    string        `json:"version"`
	APIServer  string        `json:"apiServer"`
	Nodes      NodeSummary   `json:"nodes"`
	Namespaces int           `json:"namespaces"`
	Pods       PodSummary    `json:"pods"`
	Usage      ResourceUsage `json:"usage"`
}

// GetClusterOverview 集群总览（metrics-server 不可用时降级：usage.metricsAvailable=false）
func (s *K8sClusterService) GetClusterOverview(clusterID uint) (*ClusterOverview, error) {
	cfg, c, err := restConfigForCluster(clusterID)
	if err != nil {
		return nil, err
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("构造 clientset 失败: %w", err)
	}
	ctx := context.Background()
	ov := &ClusterOverview{APIServer: c.Server}

	if sv, err := cs.Discovery().ServerVersion(); err == nil {
		ov.Version = sv.GitVersion
	}

	nodes, err := cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("节点列表获取失败: %w", err)
	}
	ov.Nodes.Total = len(nodes.Items)
	cpuAlloc, memAlloc := resource.Quantity{}, resource.Quantity{}
	for i := range nodes.Items {
		n := &nodes.Items[i]
		if isNodeReady(n) {
			ov.Nodes.Ready++
		}
		if n.Spec.Unschedulable {
			ov.Nodes.Cordoned++
		}
		cpuAlloc.Add(*n.Status.Allocatable.Cpu())
		memAlloc.Add(*n.Status.Allocatable.Memory())
	}

	pods, err := cs.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err == nil {
		ov.Pods.Total = len(pods.Items)
		for i := range pods.Items {
			if pods.Items[i].Status.Phase == "Running" {
				ov.Pods.Running++
			}
		}
	}

	nss, err := cs.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err == nil {
		ov.Namespaces = len(nss.Items)
	}

	// metrics-server 用量（不可用降级不阻断总览）
	ov.Usage.CPUAllocCores = cpuAlloc.AsApproximateFloat64()
	ov.Usage.MemAllocGB = float64(memAlloc.Value()) / (1 << 30)
	if mc, merr := metricsClientsetForCluster(cfg); merr == nil {
		if ml, lerr := mc.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{}); lerr == nil {
			ov.Usage.MetricsAvailable = true
			cpuUsed, memUsed := resource.Quantity{}, resource.Quantity{}
			for i := range ml.Items {
				cpuUsed.Add(*ml.Items[i].Usage.Cpu())
				memUsed.Add(*ml.Items[i].Usage.Memory())
			}
			ov.Usage.CPUUsedCores = cpuUsed.AsApproximateFloat64()
			ov.Usage.MemUsedGB = float64(memUsed.Value()) / (1 << 30)
			ov.Usage.CPUPercent = percentOf(ov.Usage.CPUUsedCores, ov.Usage.CPUAllocCores)
			ov.Usage.MemPercent = percentOf(ov.Usage.MemUsedGB, ov.Usage.MemAllocGB)
		}
	}
	return ov, nil
}

// percentOf 百分比（分母 0 保护）
func percentOf(used, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return used / total * 100
}
