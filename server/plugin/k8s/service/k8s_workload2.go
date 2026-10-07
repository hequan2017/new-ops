// Package service K2 补齐：StatefulSet/DaemonSet 列表、工作负载 YAML、Pod 详情/删除、PVC/Ingress/Event
package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

// StatefulSetInfo StatefulSet 列表项
type StatefulSetInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  int32  `json:"replicas"`
	Ready     int32  `json:"ready"`
	Age       string `json:"age"`
}

// ListStatefulSets StatefulSet 列表
func (s *K8sClusterService) ListStatefulSets(clusterID uint, namespace string) ([]StatefulSetInfo, error) {
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, err
	}
	list, err := cs.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("StatefulSet 列表获取失败: %w", err)
	}
	out := make([]StatefulSetInfo, 0, len(list.Items))
	for i := range list.Items {
		sts := &list.Items[i]
		out = append(out, StatefulSetInfo{
			Name: sts.Name, Namespace: sts.Namespace,
			Replicas: derefInt32(sts.Spec.Replicas), Ready: sts.Status.ReadyReplicas,
			Age: translateTimestamp(sts.CreationTimestamp.Time),
		})
	}
	return out, nil
}

// DaemonSetInfo DaemonSet 列表项
type DaemonSetInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Desired   int32  `json:"desired"`
	Ready     int32  `json:"ready"`
	Available int32  `json:"available"`
	Age       string `json:"age"`
}

// ListDaemonSets DaemonSet 列表
func (s *K8sClusterService) ListDaemonSets(clusterID uint, namespace string) ([]DaemonSetInfo, error) {
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, err
	}
	list, err := cs.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("DaemonSet 列表获取失败: %w", err)
	}
	out := make([]DaemonSetInfo, 0, len(list.Items))
	for i := range list.Items {
		ds := &list.Items[i]
		out = append(out, DaemonSetInfo{
			Name: ds.Name, Namespace: ds.Namespace,
			Desired: ds.Status.DesiredNumberScheduled, Ready: ds.Status.NumberReady,
			Available: ds.Status.NumberAvailable,
			Age:       translateTimestamp(ds.CreationTimestamp.Time),
		})
	}
	return out, nil
}

// GetWorkloadYAML 工作负载 YAML 查看（deployment/statefulset/daemonset；剔除 managedFields 噪音）
func (s *K8sClusterService) GetWorkloadYAML(clusterID uint, kind, namespace, name string) (string, error) {
	if err := validateNsName(namespace, name); err != nil {
		return "", err
	}
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	var obj runtime.Object
	switch strings.ToLower(kind) {
	case "deployment":
		obj, err = cs.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	case "statefulset":
		obj, err = cs.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	case "daemonset":
		obj, err = cs.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	default:
		return "", newK8sErr(ErrCodeKindUnsupported, fmt.Sprintf("不支持的工作负载类型: %s", kind))
	}
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", newK8sErr(ErrCodeWorkloadNotFound, fmt.Sprintf("%s 不存在: %s/%s", kind, namespace, name))
		}
		return "", fmt.Errorf("工作负载获取失败: %w", err)
	}
	if mo, ok := obj.(metav1.Object); ok {
		mo.SetManagedFields(nil)
	}
	b, err := yaml.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("YAML 序列化失败: %w", err)
	}
	return string(b), nil
}

// PodContainerInfo Pod 容器状态
type PodContainerInfo struct {
	Name     string `json:"name"`
	Image    string `json:"image"`
	Ready    bool   `json:"ready"`
	Restarts int32  `json:"restarts"`
	State    string `json:"state"`
	Reason   string `json:"reason"`
}

// PodConditionInfo Pod 条件
type PodConditionInfo struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

// PodEventInfo Pod 相关事件
type PodEventInfo struct {
	Type     string `json:"type"`
	Reason   string `json:"reason"`
	Object   string `json:"object"`
	Message  string `json:"message"`
	Count    int32  `json:"count"`
	LastTime string `json:"lastTime"`
}

// PodDetail Pod 详情
type PodDetail struct {
	Name       string             `json:"name"`
	Namespace  string             `json:"namespace"`
	Phase      string             `json:"phase"`
	Node       string             `json:"node"`
	PodIP      string             `json:"podIP"`
	HostIP     string             `json:"hostIP"`
	QoS        string             `json:"qos"`
	Restarts   int32              `json:"restarts"`
	Age        string             `json:"age"`
	Containers []PodContainerInfo `json:"containers"`
	Conditions []PodConditionInfo `json:"conditions"`
	Events     []PodEventInfo     `json:"events"`
}

// containerStateText 容器运行状态归纳（纯函数）
func containerStateText(cs corev1.ContainerStatus) (state, reason string) {
	switch {
	case cs.State.Running != nil:
		return "Running", ""
	case cs.State.Waiting != nil:
		return "Waiting", cs.State.Waiting.Reason
	case cs.State.Terminated != nil:
		return cs.State.Terminated.Reason, cs.State.Terminated.Reason
	default:
		return "Unknown", ""
	}
}

// GetPodDetail Pod 详情（基础信息+容器状态+条件+相关事件）
func (s *K8sClusterService) GetPodDetail(clusterID uint, namespace, name string) (*PodDetail, error) {
	if err := validateNsName(namespace, name); err != nil {
		return nil, err
	}
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	p, err := cs.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, newK8sErr(ErrCodeWorkloadNotFound, fmt.Sprintf("Pod 不存在: %s/%s", namespace, name))
		}
		return nil, fmt.Errorf("Pod 获取失败: %w", err)
	}
	d := &PodDetail{
		Name: p.Name, Namespace: p.Namespace,
		Phase: string(p.Status.Phase), Node: p.Spec.NodeName,
		PodIP: p.Status.PodIP, HostIP: p.Status.HostIP,
		QoS:        string(p.Status.QOSClass),
		Age:        translateTimestamp(p.CreationTimestamp.Time),
		Containers: make([]PodContainerInfo, 0, len(p.Status.ContainerStatuses)),
		Conditions: make([]PodConditionInfo, 0, len(p.Status.Conditions)),
		Events:     make([]PodEventInfo, 0, 8),
	}
	for _, cst := range p.Status.ContainerStatuses {
		state, reason := containerStateText(cst)
		d.Restarts += cst.RestartCount
		d.Containers = append(d.Containers, PodContainerInfo{
			Name: cst.Name, Image: cst.Image, Ready: cst.Ready,
			Restarts: cst.RestartCount, State: state, Reason: reason,
		})
	}
	for _, cond := range p.Status.Conditions {
		d.Conditions = append(d.Conditions, PodConditionInfo{
			Type: string(cond.Type), Status: string(cond.Status), Reason: cond.Reason,
		})
	}
	// 相关事件（fieldSelector 精确匹配 involvedObject）
	evList, err := cs.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s", name, namespace),
	})
	if err == nil {
		sort.SliceStable(evList.Items, func(i, j int) bool {
			return evList.Items[i].LastTimestamp.After(evList.Items[j].LastTimestamp.Time)
		})
		for i := range evList.Items {
			ev := &evList.Items[i]
			d.Events = append(d.Events, PodEventInfo{
				Type: ev.Type, Reason: ev.Reason,
				Object:  fmt.Sprintf("%s/%s", ev.InvolvedObject.Kind, ev.InvolvedObject.Name),
				Message: ev.Message, Count: ev.Count,
				LastTime: translateTimestamp(ev.LastTimestamp.Time),
			})
		}
	}
	return d, nil
}

// DeletePod 删除 Pod（无控制器时永久移除；有控制器将被重建）
func (s *K8sClusterService) DeletePod(clusterID uint, namespace, name string) error {
	if err := validateNsName(namespace, name); err != nil {
		return err
	}
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return err
	}
	if err := cs.CoreV1().Pods(namespace).Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return newK8sErr(ErrCodeWorkloadNotFound, fmt.Sprintf("Pod 不存在: %s/%s", namespace, name))
		}
		return fmt.Errorf("Pod 删除失败: %w", err)
	}
	return nil
}

// PVCInfo 持久卷声明列表项
type PVCInfo struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Status       string `json:"status"`
	Volume       string `json:"volume"`
	Capacity     string `json:"capacity"`
	StorageClass string `json:"storageClass"`
	Age          string `json:"age"`
}

// ListPVCs PVC 列表
func (s *K8sClusterService) ListPVCs(clusterID uint, namespace string) ([]PVCInfo, error) {
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, err
	}
	list, err := cs.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("PVC 列表获取失败: %w", err)
	}
	out := make([]PVCInfo, 0, len(list.Items))
	for i := range list.Items {
		pvc := &list.Items[i]
		capacity := "-"
		if v, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
			capacity = v.String()
		}
		sc := "-"
		if pvc.Spec.StorageClassName != nil {
			sc = *pvc.Spec.StorageClassName
		}
		out = append(out, PVCInfo{
			Name: pvc.Name, Namespace: pvc.Namespace,
			Status: string(pvc.Status.Phase), Volume: pvc.Spec.VolumeName,
			Capacity: capacity, StorageClass: sc,
			Age: translateTimestamp(pvc.CreationTimestamp.Time),
		})
	}
	return out, nil
}

// IngressInfo Ingress 列表项
type IngressInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Hosts     string `json:"hosts"`
	Paths     string `json:"paths"`
	Age       string `json:"age"`
}

// ListIngresses Ingress 列表（networking.k8s.io/v1）
func (s *K8sClusterService) ListIngresses(clusterID uint, namespace string) ([]IngressInfo, error) {
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, err
	}
	list, err := cs.NetworkingV1().Ingresses(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Ingress 列表获取失败: %w", err)
	}
	out := make([]IngressInfo, 0, len(list.Items))
	for i := range list.Items {
		ing := &list.Items[i]
		hosts := make([]string, 0, len(ing.Spec.Rules))
		paths := make([]string, 0, len(ing.Spec.Rules))
		for _, rule := range ing.Spec.Rules {
			host := rule.Host
			if host == "" {
				host = "*"
			}
			hosts = append(hosts, host)
			if rule.HTTP == nil {
				continue
			}
			for _, p := range rule.HTTP.Paths {
				svcName := p.Backend.Service.Name
				if svcName == "" {
					svcName = "-"
				}
				paths = append(paths, host+p.Path+"→"+svcName)
			}
		}
		out = append(out, IngressInfo{
			Name: ing.Name, Namespace: ing.Namespace,
			Hosts: strings.Join(hosts, ","), Paths: strings.Join(paths, " ; "),
			Age: translateTimestamp(ing.CreationTimestamp.Time),
		})
	}
	return out, nil
}

// EventInfo 集群事件列表项（全命名空间告警排查视图）
type EventInfo struct {
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Object    string `json:"object"`
	Message   string `json:"message"`
	Count     int32  `json:"count"`
	LastTime  string `json:"lastTime"`
}

// ListEvents 事件列表（按最近时间倒序，上限 200 条）
func (s *K8sClusterService) ListEvents(clusterID uint, namespace string) ([]EventInfo, error) {
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, err
	}
	list, err := cs.CoreV1().Events(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Event 列表获取失败: %w", err)
	}
	sort.SliceStable(list.Items, func(i, j int) bool {
		return list.Items[i].LastTimestamp.After(list.Items[j].LastTimestamp.Time)
	})
	limit := len(list.Items)
	if limit > 200 {
		limit = 200
	}
	out := make([]EventInfo, 0, limit)
	for i := 0; i < limit; i++ {
		ev := &list.Items[i]
		out = append(out, EventInfo{
			Namespace: ev.Namespace, Type: ev.Type, Reason: ev.Reason,
			Object:  fmt.Sprintf("%s/%s", ev.InvolvedObject.Kind, ev.InvolvedObject.Name),
			Message: ev.Message, Count: ev.Count,
			LastTime: translateTimestamp(ev.LastTimestamp.Time),
		})
	}
	return out, nil
}
