// Package service K2 资源浏览（Pod/工作负载/Node 只读 + Pod 日志）
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/k8s/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// clientsetForCluster 解密 kubeconfig 并构造 clientset（每次构造，无长连接池）
func clientsetForCluster(clusterID uint) (*kubernetes.Clientset, model.K8sCluster, error) {
	var c model.K8sCluster
	if err := global.GVA_DB.First(&c, clusterID).Error; err != nil {
		return nil, c, fmt.Errorf("集群不存在: %w", err)
	}
	kubeconfig, err := crypto.Decrypt(c.KubeconfigCipher)
	if err != nil {
		return nil, c, fmt.Errorf("kubeconfig 解密失败: %w", err)
	}
	cs, err := kubernetesNewForConfig(kubeconfig)
	if err != nil {
		return nil, c, err
	}
	return cs, c, nil
}

// kubernetesNewForConfig 从 kubeconfig 文本构造 clientset（复用解析逻辑）
func kubernetesNewForConfig(kubeconfig string) (*kubernetes.Clientset, error) {
	cfg, err := clientcmdRESTConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	return kubernetesNewForRESTConfig(cfg)
}

// PodInfo Pod 列表项（字段精简，前端表格友好）
type PodInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Restarts  int32  `json:"restarts"`
	Node      string `json:"node"`
	Age       string `json:"age"`
}

// ListPods Pod 列表（namespace 为空则全命名空间）
func (s *K8sClusterService) ListPods(clusterID uint, namespace string) ([]PodInfo, error) {
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, err
	}
	list, err := cs.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Pod 列表获取失败: %w", err)
	}
	out := make([]PodInfo, 0, len(list.Items))
	for _, p := range list.Items {
		var restarts int32
		status := "Unknown"
		for _, cs := range p.Status.ContainerStatuses {
			restarts += cs.RestartCount
			if cs.State.Running != nil {
				status = "Running"
			} else if cs.State.Waiting != nil && status == "Unknown" {
				status = cs.State.Waiting.Reason
			} else if cs.State.Terminated != nil && status == "Unknown" {
				status = "Terminated"
			}
		}
		out = append(out, PodInfo{
			Name: p.Name, Namespace: p.Namespace, Status: status,
			Restarts: restarts, Node: p.Spec.NodeName,
			Age: translateTimestamp(p.CreationTimestamp.Time),
		})
	}
	return out, nil
}

// DeploymentInfo 工作负载列表项
type DeploymentInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  int32  `json:"replicas"`
	Ready     int32  `json:"ready"`
	Available int32  `json:"available"`
	Age       string `json:"age"`
}

// ListDeployments Deployment 列表
func (s *K8sClusterService) ListDeployments(clusterID uint, namespace string) ([]DeploymentInfo, error) {
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, err
	}
	list, err := cs.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Deployment 列表获取失败: %w", err)
	}
	out := make([]DeploymentInfo, 0, len(list.Items))
	for _, d := range list.Items {
		out = append(out, DeploymentInfo{
			Name: d.Name, Namespace: d.Namespace,
			Replicas:  derefInt32(d.Spec.Replicas),
			Ready:     statusReady(d.Status.ReadyReplicas),
			Available: statusReady(d.Status.AvailableReplicas),
			Age:       translateTimestamp(d.CreationTimestamp.Time),
		})
	}
	return out, nil
}

// NodeInfo 节点列表项
type NodeInfo struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Version  string `json:"version"`
	Internal string `json:"internal"`
	Age      string `json:"age"`
}

// ListNodes 节点列表
func (s *K8sClusterService) ListNodes(clusterID uint) ([]NodeInfo, error) {
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, err
	}
	list, err := cs.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Node 列表获取失败: %w", err)
	}
	out := make([]NodeInfo, 0, len(list.Items))
	for _, n := range list.Items {
		status := "Ready"
		for _, c := range n.Status.Conditions {
			if c.Type == "Ready" && c.Status != "True" {
				status = "NotReady"
			}
		}
		out = append(out, NodeInfo{
			Name: n.Name, Status: status,
			Version:  n.Status.NodeInfo.KubeletVersion,
			Internal: internalIP(n.Status.Addresses),
			Age:      translateTimestamp(n.CreationTimestamp.Time),
		})
	}
	return out, nil
}

// GetPodLogs Pod 日志（tailLines 限流，最长 1MB）
func (s *K8sClusterService) GetPodLogs(clusterID uint, namespace, pod, container string, tailLines int64) (logs string, err error) {
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return "", err
	}
	if tailLines <= 0 || tailLines > 2000 {
		tailLines = 500
	}
	opts := &corev1.PodLogOptions{Container: container}
	tl := tailLines
	opts.TailLines = &tl
	req := cs.CoreV1().Pods(namespace).GetLogs(pod, opts)
	stream, err := req.Stream(context.Background())
	if err != nil {
		return "", fmt.Errorf("日志流打开失败: %w", err)
	}
	defer stream.Close()
	buf := make([]byte, 64*1024)
	var sb strings.Builder
	for {
		n, rerr := stream.Read(buf)
		if n > 0 {
			sb.Write(buf[:n])
			if sb.Len() > 1<<20 {
				sb.WriteString("\n...[已截断，超过 1MB]")
				break
			}
		}
		if rerr != nil {
			break
		}
	}
	return sb.String(), nil
}

func translateTimestamp(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	d := time.Since(t)
	switch {
	case d.Hours() > 24*30:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d.Hours() > 24:
		return fmt.Sprintf("%dd%dh", int(d.Hours()/24), int(d.Hours())%24)
	case d.Hours() > 1:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
}
