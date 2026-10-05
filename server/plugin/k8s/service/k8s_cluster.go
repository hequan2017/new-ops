// Package service 白泽 Kubernetes 多集群服务（M4 K1）
package service

import (
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/k8s/model"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sClusterService 集群服务
type K8sClusterService struct{}

// validateCluster 集群校验（纯逻辑）
func validateCluster(c *model.K8sCluster, kubeconfig string) error {
	if c == nil || c.Name == "" {
		return newK8sErr(ErrCodeClusterNameRequired, "集群名称不能为空")
	}
	if kubeconfig == "" {
		return newK8sErr(ErrCodeKubeconfigInvalid, "kubeconfig 内容不能为空")
	}
	return nil
}

// parseKubeconfig 解析 kubeconfig 并构造 clientset（不发起网络请求）
func parseKubeconfig(kubeconfig string) (*kubernetes.Clientset, string, error) {
	cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, "", fmt.Errorf("kubeconfig 解析失败: %w", err)
	}
	server := cfg.Host
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, server, fmt.Errorf("构造 clientset 失败: %w", err)
	}
	return cs, server, nil
}

// CreateCluster 注册集群（kubeconfig AES-256-GCM 加密落库 + 解析校验 + 连接测试）
func (s *K8sClusterService) CreateCluster(c *model.K8sCluster, kubeconfig string) error {
	if err := validateCluster(c, kubeconfig); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.K8sCluster{}).Where("name = ?", c.Name).Count(&count)
	if count > 0 {
		return newK8sErr(ErrCodeClusterDuplicate, fmt.Sprintf("集群已存在: %s", c.Name))
	}
	cs, server, err := parseKubeconfig(kubeconfig)
	if err != nil {
		return err
	}
	// 连接测试：Discovery ServerVersion（超时由 rest.Config 默认控制）
	sv, err := cs.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("集群连接测试失败: %w", err)
	}
	cipher, err := crypto.Encrypt(kubeconfig)
	if err != nil {
		return err
	}
	c.KubeconfigCipher = cipher
	c.Server = server
	c.Version = sv.GitVersion
	c.Status = model.K8sStatusOnline
	return global.GVA_DB.Create(c).Error
}

// DeleteCluster 删除集群
func (s *K8sClusterService) DeleteCluster(id uint) error {
	return global.GVA_DB.Delete(&model.K8sCluster{}, id).Error
}

// GetClusterList 集群列表（密文与 kubeconfig 均不返回，模型字段已 json:"-"）
func (s *K8sClusterService) GetClusterList(keyword string) (list []*model.K8sCluster, err error) {
	db := global.GVA_DB.Model(&model.K8sCluster{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR remark LIKE ?", kw, kw)
	}
	err = db.Order("id DESC").Find(&list).Error
	return list, err
}

// TestCluster 对已注册集群做连接测试（解密 kubeconfig → ServerVersion）
func (s *K8sClusterService) TestCluster(id uint) (version string, err error) {
	var c model.K8sCluster
	if err = global.GVA_DB.First(&c, id).Error; err != nil {
		return "", err
	}
	kubeconfig, err := crypto.Decrypt(c.KubeconfigCipher)
	if err != nil {
		return "", fmt.Errorf("kubeconfig 解密失败: %w", err)
	}
	cs, _, err := parseKubeconfig(kubeconfig)
	if err != nil {
		return "", err
	}
	sv, err := cs.Discovery().ServerVersion()
	if err != nil {
		return "", fmt.Errorf("集群连接失败: %w", err)
	}
	return sv.GitVersion, nil
}
