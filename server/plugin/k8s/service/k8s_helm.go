// Package service M8 K3 起步：Helm release 管理（列表/安装/卸载/回滚/历史，helm.sh/helm/v3 SDK）
package service

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/k8s/model"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	discoverymem "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// helmRestGetter 用集群 kubeconfig 适配 helm 的 RESTClientGetter
// （kubeconfig 原文保留：helm kube 客户端 namespace 为空时回退 ToRawKubeConfigLoader，返回 nil 会 panic）
type helmRestGetter struct {
	cfg        *rest.Config
	kubeconfig []byte
}

func (g *helmRestGetter) ToRESTConfig() (*rest.Config, error) { return g.cfg, nil }
func (g *helmRestGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(g.cfg)
	if err != nil {
		return nil, err
	}
	return discoverymem.NewMemCacheClient(dc), nil
}
func (g *helmRestGetter) ToRESTMapper() (meta.RESTMapper, error) {
	dc, err := g.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	gr, err := restmapper.GetAPIGroupResources(dc)
	if err != nil {
		return nil, err
	}
	return restmapper.NewDiscoveryRESTMapper(gr), nil
}
func (g *helmRestGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	cc, err := clientcmd.NewClientConfigFromBytes(g.kubeconfig)
	if err != nil {
		// 兜底空配置（namespace 解析失败时 helm 自行回落 default）
		return clientcmd.NewDefaultClientConfig(clientcmdapi.Config{}, &clientcmd.ConfigOverrides{})
	}
	return cc
}

// helmActionConfig 为集群构造 helm action 配置（release 存 secret，默认 namespace）
func helmActionConfig(clusterID uint, namespace string) (*action.Configuration, error) {
	if namespace == "" {
		namespace = "default"
	}
	var c model.K8sCluster
	if err := global.GVA_DB.First(&c, clusterID).Error; err != nil {
		return nil, fmt.Errorf("集群不存在: %w", err)
	}
	kubeconfig, err := crypto.Decrypt(c.KubeconfigCipher)
	if err != nil {
		return nil, fmt.Errorf("kubeconfig 解密失败: %w", err)
	}
	cfg, err := clientcmdRESTConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	actionCfg := new(action.Configuration)
	if err := actionCfg.Init(&helmRestGetter{cfg: cfg, kubeconfig: []byte(kubeconfig)}, namespace, "secret", func(_ string, _ ...interface{}) {}); err != nil {
		return nil, fmt.Errorf("helm 初始化失败: %w", err)
	}
	return actionCfg, nil
}

// HelmReleaseInfo release 列表项
type HelmReleaseInfo struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Chart      string `json:"chart"`
	AppVersion string `json:"appVersion"`
	Revision   int    `json:"revision"`
	Status     string `json:"status"`
	Updated    string `json:"updated"`
}

// ListHelmReleases release 列表（namespace 空=全命名空间）
func (s *K8sClusterService) ListHelmReleases(clusterID uint, namespace string) ([]HelmReleaseInfo, error) {
	actionCfg, err := helmActionConfig(clusterID, namespace)
	if err != nil {
		return nil, err
	}
	list := action.NewList(actionCfg)
	if namespace == "" {
		list.AllNamespaces = true
	}
	list.Deployed = true
	list.Failed = true
	list.Pending = true
	list.Uninstalled = false
	list.Uninstalling = true
	res, err := list.Run()
	if err != nil {
		return nil, fmt.Errorf("release 列表获取失败: %w", err)
	}
	out := make([]HelmReleaseInfo, 0, len(res))
	for _, r := range res {
		out = append(out, HelmReleaseInfo{
			Name: r.Name, Namespace: r.Namespace,
			Chart: chartName(r), AppVersion: appVersionOf(r),
			Revision: r.Version, Status: r.Info.Status.String(),
			Updated: translateTimestamp(r.Info.LastDeployed.Time),
		})
	}
	return out, nil
}

// InstallHelmRelease 安装/升级 release（不存在则装、存在则升——revision 递增；chart tgz 上传 + values YAML 覆盖）
func (s *K8sClusterService) InstallHelmRelease(clusterID uint, namespace, releaseName string, tgz io.Reader, valuesYAML string) (*HelmReleaseInfo, error) {
	if namespace == "" || releaseName == "" {
		return nil, newK8sErr(ErrCodeNamespaceReq, "namespace/releaseName 必填")
	}
	data, err := io.ReadAll(tgz)
	if err != nil {
		return nil, fmt.Errorf("chart 包读取失败: %w", err)
	}
	if len(data) == 0 {
		return nil, newK8sErr(ErrCodeYAMLEmpty, "chart 包不能为空")
	}
	ch, err := loader.LoadArchive(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("chart 包解析失败: %w", err)
	}
	vals := map[string]any{}
	if valuesYAML != "" {
		vals, err = chartutil.ReadValues([]byte(valuesYAML))
		if err != nil {
			return nil, fmt.Errorf("values YAML 解析失败: %w", err)
		}
	}
	actionCfg, err := helmActionConfig(clusterID, namespace)
	if err != nil {
		return nil, err
	}
	// 已存在则升级（revision 递增，回滚可用）
	probe := action.NewList(actionCfg)
	probe.Filter = "^" + releaseName + "$"
	probe.Deployed = true
	probe.Failed = true
	probe.Pending = true
	if rels, lerr := probe.Run(); lerr == nil && len(rels) > 0 {
		up := action.NewUpgrade(actionCfg)
		up.Namespace = namespace
		up.Timeout = 2 * time.Minute
		rel, uerr := up.Run(releaseName, ch, vals)
		if uerr != nil {
			return nil, fmt.Errorf("升级失败: %w", uerr)
		}
		return releaseInfo(rel), nil
	}
	install := action.NewInstall(actionCfg)
	install.ReleaseName = releaseName
	install.Namespace = namespace
	install.Timeout = 2 * time.Minute
	// 不等待资源就绪（Wait=false）：返回快，状态由 release 列表观察
	rel, err := install.Run(ch, vals)
	if err != nil {
		return nil, fmt.Errorf("安装失败: %w", err)
	}
	return releaseInfo(rel), nil
}

// UninstallHelmRelease 卸载 release
func (s *K8sClusterService) UninstallHelmRelease(clusterID uint, namespace, name string) error {
	if namespace == "" || name == "" {
		return newK8sErr(ErrCodeNamespaceReq, "namespace/name 必填")
	}
	actionCfg, err := helmActionConfig(clusterID, namespace)
	if err != nil {
		return err
	}
	un := action.NewUninstall(actionCfg)
	un.Timeout = 2 * time.Minute
	if _, err := un.Run(name); err != nil {
		return fmt.Errorf("卸载失败: %w", err)
	}
	return nil
}

// RollbackHelmRelease 回滚 release（revision<=0 回滚到上一版）
func (s *K8sClusterService) RollbackHelmRelease(clusterID uint, namespace, name string, revision int) error {
	if namespace == "" || name == "" {
		return newK8sErr(ErrCodeNamespaceReq, "namespace/name 必填")
	}
	actionCfg, err := helmActionConfig(clusterID, namespace)
	if err != nil {
		return err
	}
	if revision <= 0 {
		hist := action.NewHistory(actionCfg)
		rels, herr := hist.Run(name)
		if herr != nil {
			return fmt.Errorf("历史获取失败: %w", herr)
		}
		if len(rels) == 0 {
			return newK8sErr(ErrCodeWorkloadNotFound, fmt.Sprintf("release 不存在: %s", name))
		}
		cur := rels[len(rels)-1].Version
		revision = cur - 1
		if revision < 1 {
			return newK8sErr(ErrCodeNamespaceReq, "当前为首版，无可回滚版本")
		}
	}
	rb := action.NewRollback(actionCfg)
	rb.Version = revision
	rb.Timeout = 2 * time.Minute
	if err := rb.Run(name); err != nil {
		return fmt.Errorf("回滚失败: %w", err)
	}
	return nil
}

// HelmHistoryItem release 历史项
type HelmHistoryItem struct {
	Revision int    `json:"revision"`
	Status   string `json:"status"`
	Chart    string `json:"chart"`
	Updated  string `json:"updated"`
}

// GetHelmHistory release 历史（按 revision 升序）
func (s *K8sClusterService) GetHelmHistory(clusterID uint, namespace, name string) ([]HelmHistoryItem, error) {
	actionCfg, err := helmActionConfig(clusterID, namespace)
	if err != nil {
		return nil, err
	}
	rels, err := action.NewHistory(actionCfg).Run(name)
	if err != nil {
		return nil, fmt.Errorf("历史获取失败: %w", err)
	}
	out := make([]HelmHistoryItem, 0, len(rels))
	for _, r := range rels {
		out = append(out, HelmHistoryItem{
			Revision: r.Version,
			Status:   statusOf(r),
			Chart:    chartName(r),
			Updated:  translateTimestamp(r.Info.LastDeployed.Time),
		})
	}
	return out, nil
}

func chartName(r *release.Release) string {
	if r.Chart == nil {
		return ""
	}
	return fmt.Sprintf("%s-%s", r.Chart.Name(), r.Chart.Metadata.Version)
}

func appVersionOf(r *release.Release) string {
	if r.Chart == nil || r.Chart.Metadata == nil {
		return ""
	}
	return r.Chart.Metadata.AppVersion
}

func statusOf(r *release.Release) string {
	if r.Info == nil {
		return ""
	}
	return r.Info.Status.String()
}

func releaseInfo(r *release.Release) *HelmReleaseInfo {
	info := &HelmReleaseInfo{
		Name: r.Name, Namespace: r.Namespace,
		Chart: chartName(r), AppVersion: appVersionOf(r),
		Revision: r.Version,
	}
	if r.Info != nil {
		info.Status = r.Info.Status.String()
		info.Updated = translateTimestamp(r.Info.LastDeployed.Time)
	}
	return info
}
