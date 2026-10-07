// Package service K2 写操作：Deployment 扩缩容与滚动重启（写操作走 GVA 操作日志中间件）
package service

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// k8s 插件错误码补充（写操作）
const (
	ErrCodeReplicasInvalid  = 1505
	ErrCodeNamespaceReq     = 1506
	ErrCodeWorkloadNotFound = 1507
)

// validateReplicas 副本数校验（纯逻辑）
func validateReplicas(replicas int32) error {
	if replicas < 0 || replicas > 500 {
		return newK8sErr(ErrCodeReplicasInvalid, fmt.Sprintf("副本数超出合法范围（0-500）: %d", replicas))
	}
	return nil
}

// validateNsName 命名空间/名称校验（纯逻辑）
func validateNsName(namespace, name string) error {
	if namespace == "" {
		return newK8sErr(ErrCodeNamespaceReq, "命名空间不能为空")
	}
	if name == "" {
		return newK8sErr(ErrCodeNamespaceReq, "资源名称不能为空")
	}
	return nil
}

// ScaleDeployment 扩缩容
func (s *K8sClusterService) ScaleDeployment(clusterID uint, namespace, name string, replicas int32) error {
	if err := validateNsName(namespace, name); err != nil {
		return err
	}
	if err := validateReplicas(replicas); err != nil {
		return err
	}
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return err
	}
	deploy, err := cs.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return newK8sErr(ErrCodeWorkloadNotFound, fmt.Sprintf("Deployment 不存在: %s/%s", namespace, name))
	}
	deploy.Spec.Replicas = &replicas
	_, err = cs.AppsV1().Deployments(namespace).Update(context.Background(), deploy, metav1.UpdateOptions{})
	return err
}

// timeNowUTC 时间源（测试可替换）
var timeNowUTC = func() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

// RestartDeployment 滚动重启（patch restartedAt 注解触发滚动）
func (s *K8sClusterService) RestartDeployment(clusterID uint, namespace, name string) error {
	if err := validateNsName(namespace, name); err != nil {
		return err
	}
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return err
	}
	payload := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, timeNowUTC())
	_, err = cs.AppsV1().Deployments(namespace).Patch(context.Background(), name,
		types.StrategicMergePatchType, []byte(payload), metav1.PatchOptions{})
	return err
}
