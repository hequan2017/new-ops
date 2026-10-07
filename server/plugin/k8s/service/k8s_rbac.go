// Package service 三级 RBAC 之命名空间层：普通用户仅见授权 namespace（888 全量不经此表；
// 写操作与终端仍按 casbin 仅 888——本层只做浏览面隔离）
package service

import (
	"context"
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/k8s/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// isSuperAdmin 超管判定（888）
func isSuperAdmin(authorityID uint) bool { return authorityID == 888 }

// grantedNamespaces 用户在集群被授权的 namespace 列表
func grantedNamespaces(clusterID, userID uint) ([]string, error) {
	var rows []model.K8sNsGrant
	if err := global.GVA_DB.Where("cluster_id = ? AND user_id = ?", clusterID, userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Namespace)
	}
	return out, nil
}

// namespaceVisible 指定 namespace 对用户是否可见（超管恒真）
func namespaceVisible(clusterID, userID, authorityID uint, namespace string) bool {
	if isSuperAdmin(authorityID) {
		return true
	}
	granted, err := grantedNamespaces(clusterID, userID)
	if err != nil {
		return false
	}
	for _, ns := range granted {
		if ns == namespace {
			return true
		}
	}
	return false
}

// scopedList 命名空间类列表统一过滤：超管按原语义（空=全部）；
// 普通用户显式 ns 须在授权内（否则空结果），空 ns 聚合全部授权 ns
func scopedList[T any](clusterID, userID, authorityID uint, namespace string, fetch func(ns string) ([]T, error)) ([]T, error) {
	if isSuperAdmin(authorityID) {
		return fetch(namespace)
	}
	granted, err := grantedNamespaces(clusterID, userID)
	if err != nil {
		return nil, err
	}
	if namespace != "" {
		for _, ns := range granted {
			if ns == namespace {
				return fetch(namespace)
			}
		}
		return []T{}, nil
	}
	var out []T
	for _, ns := range granted {
		part, err := fetch(ns)
		if err != nil {
			return nil, err
		}
		out = append(out, part...)
	}
	if out == nil {
		out = []T{}
	}
	return out, nil
}

// NsVisibilityItem 命名空间可见性（前端下拉与授权管理共用）
type NsVisibilityItem struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Granted bool   `json:"granted"` // 对当前用户是否已授权（超管恒真）
}

// ListNamespacesForUser 集群实际 namespace 列表 + 当前用户授权标记
func (s *K8sClusterService) ListNamespacesForUser(clusterID, userID, authorityID uint) ([]NsVisibilityItem, error) {
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, err
	}
	nsList, err := cs.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("命名空间列表获取失败: %w", err)
	}
	grantedSet := map[string]bool{}
	if !isSuperAdmin(authorityID) {
		granted, err := grantedNamespaces(clusterID, userID)
		if err != nil {
			return nil, err
		}
		for _, ns := range granted {
			grantedSet[ns] = true
		}
	}
	out := make([]NsVisibilityItem, 0, len(nsList.Items))
	for i := range nsList.Items {
		ns := &nsList.Items[i]
		out = append(out, NsVisibilityItem{
			Name:    ns.Name,
			Status:  string(ns.Status.Phase),
			Granted: isSuperAdmin(authorityID) || grantedSet[ns.Name],
		})
	}
	return out, nil
}

// CreateNsGrant 授予命名空间可见性（幂等）
func (s *K8sClusterService) CreateNsGrant(g *model.K8sNsGrant) error {
	if g.ClusterID == 0 || g.Namespace == "" || g.UserID == 0 {
		return newK8sErr(ErrCodeNamespaceReq, "clusterId/namespace/userId 必填")
	}
	var count int64
	global.GVA_DB.Model(&model.K8sNsGrant{}).
		Where("cluster_id = ? AND namespace = ? AND user_id = ?", g.ClusterID, g.Namespace, g.UserID).Count(&count)
	if count > 0 {
		return newK8sErr(ErrCodeClusterDuplicate, "该用户已被授权此命名空间")
	}
	return global.GVA_DB.Create(g).Error
}

// DeleteNsGrant 收回授权
func (s *K8sClusterService) DeleteNsGrant(id uint) error {
	return global.GVA_DB.Delete(&model.K8sNsGrant{}, id).Error
}

// ListNsGrants 集群授权清单（含用户名，888 管理面）
type NsGrantItem struct {
	ID        uint   `json:"ID"`
	ClusterID uint   `json:"clusterId"`
	Namespace string `json:"namespace"`
	UserID    uint   `json:"userId"`
	Username  string `json:"username"`
	NickName  string `json:"nickName"`
}

func (s *K8sClusterService) ListNsGrants(clusterID uint) ([]NsGrantItem, error) {
	var rows []model.K8sNsGrant
	if err := global.GVA_DB.Where("cluster_id = ?", clusterID).Order("namespace, user_id").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]NsGrantItem, 0, len(rows))
	for _, r := range rows {
		item := NsGrantItem{ID: r.ID, ClusterID: r.ClusterID, Namespace: r.Namespace, UserID: r.UserID}
		// 用户信息联查（独立结构承接——Scan 会按列覆写同名映射字段，不可扫进预填结构）
		var u struct {
			Username string `gorm:"column:username"`
			NickName string `gorm:"column:nick_name"`
		}
		global.GVA_DB.Table("sys_users").Select("username, nick_name").Where("id = ?", r.UserID).Scan(&u)
		item.Username, item.NickName = u.Username, u.NickName
		out = append(out, item)
	}
	return out, nil
}

// GuardNamespace 单对象读写守卫（detail/logs/yaml/history 等命名空间内资源）
func (s *K8sClusterService) GuardNamespace(clusterID, userID, authorityID uint, namespace string) error {
	if namespaceVisible(clusterID, userID, authorityID, namespace) {
		return nil
	}
	return newK8sErr(ErrCodeNsForbidden, fmt.Sprintf("无命名空间 %s 的访问权限", namespace))
}

// ScopedOverview 总览的浏览面隔离：普通用户 namespaces/pods 计数按授权 ns 聚合（节点/用量为集群基础设施面保持只读汇总）
func (s *K8sClusterService) ScopedOverview(clusterID, userID, authorityID uint) (*ClusterOverview, error) {
	ov, err := s.GetClusterOverview(clusterID)
	if err != nil {
		return nil, err
	}
	if isSuperAdmin(authorityID) {
		return ov, nil
	}
	granted, err := grantedNamespaces(clusterID, userID)
	if err != nil {
		return nil, err
	}
	ov.Namespaces = len(granted)
	if len(granted) == 0 {
		ov.Pods = PodSummary{}
		return ov, nil
	}
	cs, _, err := clientsetForCluster(clusterID)
	if err != nil {
		return nil, err
	}
	total, running := 0, 0
	for _, ns := range granted {
		pods, err := cs.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			continue
		}
		total += len(pods.Items)
		for i := range pods.Items {
			if pods.Items[i].Status.Phase == "Running" {
				running++
			}
		}
	}
	ov.Pods = PodSummary{Total: total, Running: running}
	return ov, nil
}

// ---- 列表类 scoped 变体（委托原实现 + 本文件过滤） ----

func (s *K8sClusterService) ListPodsScoped(clusterID uint, namespace string, userID, authorityID uint) ([]PodInfo, error) {
	return scopedList(clusterID, userID, authorityID, namespace, func(ns string) ([]PodInfo, error) {
		return s.ListPods(clusterID, ns)
	})
}

func (s *K8sClusterService) ListDeploymentsScoped(clusterID uint, namespace string, userID, authorityID uint) ([]DeploymentInfo, error) {
	return scopedList(clusterID, userID, authorityID, namespace, func(ns string) ([]DeploymentInfo, error) {
		return s.ListDeployments(clusterID, ns)
	})
}

func (s *K8sClusterService) ListStatefulSetsScoped(clusterID uint, namespace string, userID, authorityID uint) ([]StatefulSetInfo, error) {
	return scopedList(clusterID, userID, authorityID, namespace, func(ns string) ([]StatefulSetInfo, error) {
		return s.ListStatefulSets(clusterID, ns)
	})
}

func (s *K8sClusterService) ListDaemonSetsScoped(clusterID uint, namespace string, userID, authorityID uint) ([]DaemonSetInfo, error) {
	return scopedList(clusterID, userID, authorityID, namespace, func(ns string) ([]DaemonSetInfo, error) {
		return s.ListDaemonSets(clusterID, ns)
	})
}

func (s *K8sClusterService) ListServicesScoped(clusterID uint, namespace string, userID, authorityID uint) ([]ServiceInfo, error) {
	return scopedList(clusterID, userID, authorityID, namespace, func(ns string) ([]ServiceInfo, error) {
		return s.ListServices(clusterID, ns)
	})
}

func (s *K8sClusterService) ListConfigMapsScoped(clusterID uint, namespace string, userID, authorityID uint) ([]ConfigMapInfo, error) {
	return scopedList(clusterID, userID, authorityID, namespace, func(ns string) ([]ConfigMapInfo, error) {
		return s.ListConfigMaps(clusterID, ns)
	})
}

func (s *K8sClusterService) ListSecretsScoped(clusterID uint, namespace string, userID, authorityID uint) ([]SecretInfo, error) {
	return scopedList(clusterID, userID, authorityID, namespace, func(ns string) ([]SecretInfo, error) {
		return s.ListSecrets(clusterID, ns)
	})
}

func (s *K8sClusterService) ListPVCsScoped(clusterID uint, namespace string, userID, authorityID uint) ([]PVCInfo, error) {
	return scopedList(clusterID, userID, authorityID, namespace, func(ns string) ([]PVCInfo, error) {
		return s.ListPVCs(clusterID, ns)
	})
}

func (s *K8sClusterService) ListIngressesScoped(clusterID uint, namespace string, userID, authorityID uint) ([]IngressInfo, error) {
	return scopedList(clusterID, userID, authorityID, namespace, func(ns string) ([]IngressInfo, error) {
		return s.ListIngresses(clusterID, ns)
	})
}

func (s *K8sClusterService) ListEventsScoped(clusterID uint, namespace string, userID, authorityID uint) ([]EventInfo, error) {
	return scopedList(clusterID, userID, authorityID, namespace, func(ns string) ([]EventInfo, error) {
		return s.ListEvents(clusterID, ns)
	})
}

func (s *K8sClusterService) ListHelmScoped(clusterID uint, namespace string, userID, authorityID uint) ([]HelmReleaseInfo, error) {
	return scopedList(clusterID, userID, authorityID, namespace, func(ns string) ([]HelmReleaseInfo, error) {
		return s.ListHelmReleases(clusterID, ns)
	})
}
