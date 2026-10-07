// Package service K1 Node 管理（详情/cordon/uncordon/drain）
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// NodeResource CPU/内存/Pod 配额三元组（格式化字符串，前端直显）
type NodeResource struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	Pods   string `json:"pods"`
}

// NodeConditionInfo 节点条件
type NodeConditionInfo struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// NodeTaintInfo 节点污点
type NodeTaintInfo struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

// NodeDetail 节点详情
type NodeDetail struct {
	Name          string              `json:"name"`
	Status        string              `json:"status"`
	Roles         string              `json:"roles"`
	Kubelet       string              `json:"kubelet"`
	InternalIP    string              `json:"internalIP"`
	OSImage       string              `json:"osImage"`
	Arch          string              `json:"arch"`
	Unschedulable bool                `json:"unschedulable"`
	Age           string              `json:"age"`
	Allocatable   NodeResource        `json:"allocatable"`
	Capacity      NodeResource        `json:"capacity"`
	Conditions    []NodeConditionInfo `json:"conditions"`
	Taints        []NodeTaintInfo     `json:"taints"`
}

// summarizeNodeStatus 节点状态归纳（纯函数）：Ready/NotReady 与 SchedulingDisabled 组合
func summarizeNodeStatus(n *corev1.Node) string {
	ready := false
	for _, c := range n.Status.Conditions {
		if c.Type == corev1.NodeReady {
			ready = c.Status == corev1.ConditionTrue
		}
	}
	status := "NotReady"
	if ready {
		status = "Ready"
	}
	if n.Spec.Unschedulable {
		status += ",SchedulingDisabled"
	}
	return status
}

// isNodeReady 节点 Ready 条件判定
func isNodeReady(n *corev1.Node) bool {
	for _, c := range n.Status.Conditions {
		if c.Type == corev1.NodeReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

// nodeRoles 节点角色（label node-role.kubernetes.io/*）
func nodeRoles(n *corev1.Node) string {
	roles := make([]string, 0, 2)
	for k := range n.Labels {
		if strings.HasPrefix(k, "node-role.kubernetes.io/") {
			roles = append(roles, strings.TrimPrefix(k, "node-role.kubernetes.io/"))
		}
	}
	if len(roles) == 0 {
		return "-"
	}
	return strings.Join(roles, ",")
}

// quantityOrDash Quantity 格式化（nil 保护）
func quantityOrDash(q *resource.Quantity) string {
	if q == nil {
		return "-"
	}
	return q.String()
}

// nodeResourceFrom 提取 CPU/内存/Pod 三元组
func nodeResourceFrom(list corev1.ResourceList) NodeResource {
	return NodeResource{
		CPU:    quantityOrDash(list.Cpu()),
		Memory: quantityOrDash(list.Memory()),
		Pods:   quantityOrDash(list.Pods()),
	}
}

// getNodeAndClientset 取节点对象与 clientset（详情/写操作共用）
func getNodeAndClientset(clusterID uint, name string) (*corev1.Node, *kubernetes.Clientset, error) {
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, nil, err
	}
	n, err := cs.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil, newK8sErr(ErrCodeNodeNotFound, fmt.Sprintf("节点不存在: %s", name))
		}
		return nil, nil, fmt.Errorf("节点获取失败: %w", err)
	}
	return n, cs, nil
}

// GetNodeDetail 节点详情
func (s *K8sClusterService) GetNodeDetail(clusterID uint, name string) (*NodeDetail, error) {
	n, _, err := getNodeAndClientset(clusterID, name)
	if err != nil {
		return nil, err
	}
	d := &NodeDetail{
		Name:          n.Name,
		Status:        summarizeNodeStatus(n),
		Roles:         nodeRoles(n),
		Kubelet:       n.Status.NodeInfo.KubeletVersion,
		InternalIP:    internalIP(n.Status.Addresses),
		OSImage:       n.Status.NodeInfo.OSImage,
		Arch:          n.Status.NodeInfo.Architecture,
		Unschedulable: n.Spec.Unschedulable,
		Age:           translateTimestamp(n.CreationTimestamp.Time),
		Allocatable:   nodeResourceFrom(n.Status.Allocatable),
		Capacity:      nodeResourceFrom(n.Status.Capacity),
		Conditions:    make([]NodeConditionInfo, 0, len(n.Status.Conditions)),
		Taints:        make([]NodeTaintInfo, 0, len(n.Spec.Taints)),
	}
	for _, c := range n.Status.Conditions {
		d.Conditions = append(d.Conditions, NodeConditionInfo{
			Type: string(c.Type), Status: string(c.Status), Reason: c.Reason, Message: c.Message,
		})
	}
	for _, t := range n.Spec.Taints {
		d.Taints = append(d.Taints, NodeTaintInfo{Key: t.Key, Value: t.Value, Effect: string(t.Effect)})
	}
	return d, nil
}

// cordonPatch cordon/uncordon 的 strategic merge patch（纯函数，可单测）
func cordonPatch(cordon bool) []byte {
	body := map[string]any{
		"spec": map[string]any{"unschedulable": cordon},
	}
	b, _ := json.Marshal(body)
	return b
}

// SetNodeSchedulability 节点隔离/恢复调度（cordon=true 隔离）
func (s *K8sClusterService) SetNodeSchedulability(clusterID uint, name string, cordon bool) error {
	_, cs, err := getNodeAndClientset(clusterID, name)
	if err != nil {
		return err
	}
	if _, err := cs.CoreV1().Nodes().Patch(context.Background(), name, types.StrategicMergePatchType, cordonPatch(cordon), metav1.PatchOptions{}); err != nil {
		return fmt.Errorf("节点调度状态更新失败: %w", err)
	}
	return nil
}

// drainItem 驱逐候选/跳过明细
type drainItem struct {
	Namespace  string `json:"namespace"`
	Pod        string `json:"pod"`
	Controller string `json:"controller,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

// DrainResult 驱逐结果汇总
type DrainResult struct {
	Node     string      `json:"node"`
	Cordoned bool        `json:"cordoned"`
	Evicted  []drainItem `json:"evicted"`
	Skipped  []drainItem `json:"skipped"`
	Failed   []drainItem `json:"failed"`
}

// podControllerKind 取 Pod 属主控制器类型（最高优先 OwnerReference）
func podControllerKind(p *corev1.Pod) string {
	for _, or := range p.OwnerReferences {
		if or.Controller != nil && *or.Controller {
			return or.Kind
		}
	}
	return ""
}

// drainDecisions 驱逐决策（纯函数）：DaemonSet 与镜像 Pod 跳过，其余进驱逐队列
func drainDecisions(pods []corev1.Pod) (evict, skip []drainItem) {
	for i := range pods {
		p := &pods[i]
		item := drainItem{Namespace: p.Namespace, Pod: p.Name, Controller: podControllerKind(p)}
		switch {
		case p.Annotations["kubernetes.io/config.mirror"] != "":
			item.Reason = "静态镜像 Pod"
			skip = append(skip, item)
		case item.Controller == "DaemonSet":
			item.Reason = "DaemonSet Pod"
			skip = append(skip, item)
		case p.DeletionTimestamp != nil:
			item.Reason = "删除中"
			skip = append(skip, item)
		default:
			if item.Controller == "" {
				item.Controller = "独立Pod"
			}
			evict = append(evict, item)
		}
	}
	return evict, skip
}

// DrainNode 节点驱逐：先 cordon，再对非 DaemonSet/非镜像 Pod 提交 Eviction（gracePeriod 秒，<0 用 Pod 默认）
func (s *K8sClusterService) DrainNode(clusterID uint, name string, gracePeriod int64) (*DrainResult, error) {
	_, cs, err := getNodeAndClientset(clusterID, name)
	if err != nil {
		return nil, err
	}
	if _, err := cs.CoreV1().Nodes().Patch(context.Background(), name, types.StrategicMergePatchType, cordonPatch(true), metav1.PatchOptions{}); err != nil {
		return nil, fmt.Errorf("驱逐前隔离失败: %w", err)
	}

	pods, err := cs.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + name,
	})
	if err != nil {
		return nil, fmt.Errorf("节点 Pod 列表获取失败: %w", err)
	}

	res := &DrainResult{Node: name, Cordoned: true}
	evict, skip := drainDecisions(pods.Items)
	res.Skipped = skip
	ctx := context.Background()
	for _, item := range evict {
		delOpts := &metav1.DeleteOptions{}
		if gracePeriod >= 0 {
			delOpts.GracePeriodSeconds = &gracePeriod
		}
		ev := &policyv1.Eviction{
			ObjectMeta:    metav1.ObjectMeta{Namespace: item.Namespace, Name: item.Pod},
			DeleteOptions: delOpts,
		}
		if err := cs.PolicyV1().Evictions(item.Namespace).Evict(ctx, ev); err != nil {
			item.Reason = err.Error()
			res.Failed = append(res.Failed, item)
			continue
		}
		res.Evicted = append(res.Evicted, item)
	}
	return res, nil
}
