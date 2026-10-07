// Package service K2 收官：工作负载 YAML 编辑下发（解析校验 / server dry-run diff 预览 / Apply）
package service

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// diffLines 行级 diff（LCS 动态规划；纯函数，可单测）
// 返回带前缀的行：两空格=相同 / -=旧 / +=新；超限返回提示占位
func diffLines(a, b []string) []string {
	const maxLines = 1200
	if len(a) > maxLines || len(b) > maxLines {
		return []string{fmt.Sprintf("（内容超过 %d 行，跳过逐行对比）", maxLines)}
	}
	n, m := len(a), len(b)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	out := make([]string, 0, n+m)
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			out = append(out, "  "+a[i])
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			out = append(out, "- "+a[i])
			i++
		default:
			out = append(out, "+ "+b[j])
			j++
		}
	}
	for ; i < n; i++ {
		out = append(out, "- "+a[i])
	}
	for ; j < m; j++ {
		out = append(out, "+ "+b[j])
	}
	return out
}

// parseTypedWorkload 按类型解析 YAML 为强类型对象
func parseTypedWorkload(kind string, data []byte) (runtime.Object, metav1.Object, error) {
	switch strings.ToLower(kind) {
	case "deployment":
		var d appsv1.Deployment
		if err := yaml.Unmarshal(data, &d); err != nil {
			return nil, nil, fmt.Errorf("Deployment YAML 解析失败: %w", err)
		}
		return &d, &d, nil
	case "statefulset":
		var s appsv1.StatefulSet
		if err := yaml.Unmarshal(data, &s); err != nil {
			return nil, nil, fmt.Errorf("StatefulSet YAML 解析失败: %w", err)
		}
		return &s, &s, nil
	case "daemonset":
		var ds appsv1.DaemonSet
		if err := yaml.Unmarshal(data, &ds); err != nil {
			return nil, nil, fmt.Errorf("DaemonSet YAML 解析失败: %w", err)
		}
		return &ds, &ds, nil
	default:
		return nil, nil, newK8sErr(ErrCodeKindUnsupported, fmt.Sprintf("不支持的工作负载类型: %s", kind))
	}
}

// validateWorkloadIdentity 校验 YAML 对象与目标一致（防跨对象写，纯逻辑）
func validateWorkloadIdentity(mo metav1.Object, namespace, name string) error {
	if mo.GetName() != name || mo.GetNamespace() != namespace {
		return newK8sErr(ErrCodeYAMLIdentityMismatch, fmt.Sprintf(
			"YAML 对象与目标不一致: %s/%s（期望 %s/%s）", mo.GetNamespace(), mo.GetName(), namespace, name))
	}
	return nil
}

// stripAndMarshal 序列化工作负载（剔 managedFields/creationTimestamp 噪音，保证 diff 只显实质变化）
func stripAndMarshal(obj runtime.Object) ([]string, error) {
	if mo, ok := obj.(metav1.Object); ok {
		mo.SetManagedFields(nil)
		mo.SetCreationTimestamp(metav1.Time{})
	}
	b, err := yaml.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(b), "\n"), nil
}

// workloadOps 按类型分派的 get/update 闭包对
type workloadOps struct {
	get    func(ctx context.Context) (runtime.Object, error)
	update func(ctx context.Context, obj runtime.Object, dryRun []string) (runtime.Object, error)
}

// workloadOpsFor 按类型构造 get/update 分派
func workloadOpsFor(cs *kubernetes.Clientset, kind, namespace, name string) (*workloadOps, error) {
	switch strings.ToLower(kind) {
	case "deployment":
		return &workloadOps{
			get: func(ctx context.Context) (runtime.Object, error) {
				return cs.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
			},
			update: func(ctx context.Context, obj runtime.Object, dryRun []string) (runtime.Object, error) {
				return cs.AppsV1().Deployments(namespace).Update(ctx, obj.(*appsv1.Deployment), metav1.UpdateOptions{DryRun: dryRun})
			},
		}, nil
	case "statefulset":
		return &workloadOps{
			get: func(ctx context.Context) (runtime.Object, error) {
				return cs.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
			},
			update: func(ctx context.Context, obj runtime.Object, dryRun []string) (runtime.Object, error) {
				return cs.AppsV1().StatefulSets(namespace).Update(ctx, obj.(*appsv1.StatefulSet), metav1.UpdateOptions{DryRun: dryRun})
			},
		}, nil
	case "daemonset":
		return &workloadOps{
			get: func(ctx context.Context) (runtime.Object, error) {
				return cs.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
			},
			update: func(ctx context.Context, obj runtime.Object, dryRun []string) (runtime.Object, error) {
				return cs.AppsV1().DaemonSets(namespace).Update(ctx, obj.(*appsv1.DaemonSet), metav1.UpdateOptions{DryRun: dryRun})
			},
		}, nil
	default:
		return nil, newK8sErr(ErrCodeKindUnsupported, fmt.Sprintf("不支持的工作负载类型: %s", kind))
	}
}

// prepareApply 公共前置：clientset、类型分派、当前对象、解析与一致性校验
func (s *K8sClusterService) prepareApply(clusterID uint, kind, namespace, name, yamlText string) (*workloadOps, runtime.Object, runtime.Object, error) {
	if err := validateNsName(namespace, name); err != nil {
		return nil, nil, nil, err
	}
	if strings.TrimSpace(yamlText) == "" {
		return nil, nil, nil, newK8sErr(ErrCodeYAMLEmpty, "YAML 内容不能为空")
	}
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, nil, nil, err
	}
	ops, err := workloadOpsFor(cs, kind, namespace, name)
	if err != nil {
		return nil, nil, nil, err
	}
	cur, err := ops.get(context.Background())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("当前对象获取失败: %w", err)
	}
	newObj, newMeta, err := parseTypedWorkload(kind, []byte(yamlText))
	if err != nil {
		return nil, nil, nil, err
	}
	if err := validateWorkloadIdentity(newMeta, namespace, name); err != nil {
		return nil, nil, nil, err
	}
	return ops, cur, newObj, nil
}

// PreviewWorkloadYAML 服务端 dry-run 试算并返回与当前对象的行 diff
func (s *K8sClusterService) PreviewWorkloadYAML(clusterID uint, kind, namespace, name, yamlText string) (string, error) {
	ops, cur, newObj, err := s.prepareApply(clusterID, kind, namespace, name, yamlText)
	if err != nil {
		return "", err
	}
	after, err := ops.update(context.Background(), newObj, []string{metav1.DryRunAll})
	if err != nil {
		return "", fmt.Errorf("服务端试算失败: %w", err)
	}
	curLines, err := stripAndMarshal(cur)
	if err != nil {
		return "", err
	}
	afterLines, err := stripAndMarshal(after)
	if err != nil {
		return "", err
	}
	diff := diffLines(curLines, afterLines)
	changed := 0
	for _, l := range diff {
		if strings.HasPrefix(l, "+") || strings.HasPrefix(l, "-") {
			changed++
		}
	}
	if changed == 0 {
		return "（无实质变更）", nil
	}
	return strings.Join(diff, "\n"), nil
}

// ApplyWorkloadYAML 实际下发（乐观锁：携带当前 resourceVersion，冲突由 API server 拒绝）
func (s *K8sClusterService) ApplyWorkloadYAML(clusterID uint, kind, namespace, name, yamlText string) error {
	ops, cur, newObj, err := s.prepareApply(clusterID, kind, namespace, name, yamlText)
	if err != nil {
		return err
	}
	curMeta, cok := cur.(metav1.Object)
	newMeta, nok := newObj.(metav1.Object)
	if cok && nok {
		newMeta.SetResourceVersion(curMeta.GetResourceVersion())
		newMeta.SetUID(curMeta.GetUID())
	}
	if _, err := ops.update(context.Background(), newObj, nil); err != nil {
		return fmt.Errorf("YAML 下发失败: %w", err)
	}
	return nil
}
