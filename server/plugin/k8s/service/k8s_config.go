// Package service K2 补充：Service/ConfigMap/Secret 只读列表
package service

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceInfo k8s Service 列表项
type ServiceInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	ClusterIP string `json:"clusterIp"`
	Ports     string `json:"ports"`
	Age       string `json:"age"`
}

// ListServices Service 列表
func (s *K8sClusterService) ListServices(clusterID uint, namespace string) ([]ServiceInfo, error) {
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, err
	}
	list, err := cs.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Service 列表获取失败: %w", err)
	}
	out := make([]ServiceInfo, 0, len(list.Items))
	for _, sv := range list.Items {
		ports := make([]string, 0, len(sv.Spec.Ports))
		for _, p := range sv.Spec.Ports {
			ports = append(ports, fmt.Sprintf("%d/%s", p.Port, p.Protocol))
		}
		out = append(out, ServiceInfo{
			Name: sv.Name, Namespace: sv.Namespace,
			Type: string(sv.Spec.Type), ClusterIP: sv.Spec.ClusterIP,
			Ports: strings.Join(ports, ","), Age: translateTimestamp(sv.CreationTimestamp.Time),
		})
	}
	return out, nil
}

// ConfigMapInfo ConfigMap 列表项
type ConfigMapInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	DataKeys  string `json:"dataKeys"`
	Age       string `json:"age"`
}

// ListConfigMaps ConfigMap 列表（只回显数据键，不回显值）
func (s *K8sClusterService) ListConfigMaps(clusterID uint, namespace string) ([]ConfigMapInfo, error) {
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, err
	}
	list, err := cs.CoreV1().ConfigMaps(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("ConfigMap 列表获取失败: %w", err)
	}
	out := make([]ConfigMapInfo, 0, len(list.Items))
	for _, cm := range list.Items {
		keys := make([]string, 0, len(cm.Data))
		for k := range cm.Data {
			keys = append(keys, k)
		}
		out = append(out, ConfigMapInfo{
			Name: cm.Name, Namespace: cm.Namespace,
			DataKeys: strings.Join(keys, ","), Age: translateTimestamp(cm.CreationTimestamp.Time),
		})
	}
	return out, nil
}

// SecretInfo Secret 列表项（只回显类型与键名，值永不回显）
type SecretInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	DataKeys  string `json:"dataKeys"`
	Age       string `json:"age"`
}

// ListSecrets Secret 列表（值永不回显）
func (s *K8sClusterService) ListSecrets(clusterID uint, namespace string) ([]SecretInfo, error) {
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, err
	}
	list, err := cs.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Secret 列表获取失败: %w", err)
	}
	out := make([]SecretInfo, 0, len(list.Items))
	for _, sec := range list.Items {
		keys := make([]string, 0, len(sec.Data))
		for k := range sec.Data {
			keys = append(keys, k)
		}
		out = append(out, SecretInfo{
			Name: sec.Name, Namespace: sec.Namespace,
			Type: string(sec.Type), DataKeys: strings.Join(keys, ","), Age: translateTimestamp(sec.CreationTimestamp.Time),
		})
	}
	return out, nil
}
