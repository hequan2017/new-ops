// Package service M8 K3 起步：Helm release 管理（列表/安装/卸载/回滚/历史，helm.sh/helm/v3 SDK）
package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/k8s/model"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	discoverymem "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/yaml"
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
	ch, vals, err := loadChartAndValues(tgz, valuesYAML)
	if err != nil {
		return nil, err
	}
	actionCfg, err := helmActionConfig(clusterID, namespace)
	if err != nil {
		return nil, err
	}
	return installOrUpgrade(actionCfg, namespace, releaseName, ch, vals)
}

// loadChartAndValues tgz 读取 + values YAML 解析（安装双路径共用）
func loadChartAndValues(tgz io.Reader, valuesYAML string) (*chart.Chart, map[string]any, error) {
	data, err := io.ReadAll(tgz)
	if err != nil {
		return nil, nil, fmt.Errorf("chart 包读取失败: %w", err)
	}
	if len(data) == 0 {
		return nil, nil, newK8sErr(ErrCodeYAMLEmpty, "chart 包不能为空")
	}
	ch, err := loader.LoadArchive(bytes.NewReader(data))
	if err != nil {
		return nil, nil, fmt.Errorf("chart 包解析失败: %w", err)
	}
	vals, err := parseValuesYAML(valuesYAML)
	if err != nil {
		return nil, nil, err
	}
	return ch, vals, nil
}

func parseValuesYAML(valuesYAML string) (map[string]any, error) {
	if valuesYAML == "" {
		return map[string]any{}, nil
	}
	vals, err := chartutil.ReadValues([]byte(valuesYAML))
	if err != nil {
		return nil, fmt.Errorf("values YAML 解析失败: %w", err)
	}
	return vals, nil
}

// installOrUpgrade 同名已存在走 upgrade（revision 递增），否则 install；不等待就绪快速返回
func installOrUpgrade(actionCfg *action.Configuration, namespace, releaseName string, ch *chart.Chart, vals map[string]any) (*HelmReleaseInfo, error) {
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
	rel, err := install.Run(ch, vals)
	if err != nil {
		return nil, fmt.Errorf("安装失败: %w", err)
	}
	return releaseInfo(rel), nil
}

// validateRepoURL 仓库地址校验（仅 http/https）
func validateRepoURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return newK8sErr(ErrCodeRepoURLInvalid, "仓库地址无效")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return newK8sErr(ErrCodeRepoURLInvalid, "仓库地址仅支持 http/https")
	}
	if u.Host == "" {
		return newK8sErr(ErrCodeRepoURLInvalid, "仓库地址缺少主机")
	}
	return nil
}

// CreateHelmRepo 登记 chart 仓库
func (s *K8sClusterService) CreateHelmRepo(r *model.K8sHelmRepo) error {
	r.Name = strings.TrimSpace(r.Name)
	r.URL = strings.TrimSpace(r.URL)
	if r.Name == "" {
		return newK8sErr(ErrCodeNamespaceReq, "仓库名不能为空")
	}
	if err := validateRepoURL(r.URL); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.K8sHelmRepo{}).Where("name = ?", r.Name).Count(&count)
	if count > 0 {
		return newK8sErr(ErrCodeClusterDuplicate, fmt.Sprintf("仓库已存在: %s", r.Name))
	}
	return global.GVA_DB.Create(r).Error
}

// DeleteHelmRepo 删除仓库登记
func (s *K8sClusterService) DeleteHelmRepo(id uint) error {
	return global.GVA_DB.Delete(&model.K8sHelmRepo{}, id).Error
}

// GetHelmRepoList 仓库列表
func (s *K8sClusterService) GetHelmRepoList() ([]*model.K8sHelmRepo, error) {
	var list []*model.K8sHelmRepo
	err := global.GVA_DB.Order("id DESC").Find(&list).Error
	return list, err
}

// locateRepoChart 经 downloader 拉取仓库 chart 到本地缓存后加载（RepoURL 直连，无需预登记 repositories.yaml）
func locateRepoChart(repo *model.K8sHelmRepo, chartRef, version string) (*chart.Chart, error) {
	cpo := &action.ChartPathOptions{RepoURL: repo.URL, Version: strings.TrimSpace(version)}
	settings := cli.New()
	if settings.RepositoryCache == "" {
		settings.RepositoryCache = filepath.Join(os.TempDir(), "helm-cache")
	}
	chartPath, err := cpo.LocateChart(chartRef, settings)
	if err != nil {
		return nil, fmt.Errorf("chart 定位失败（仓库 %s）: %w", repo.Name, err)
	}
	ch, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("chart 加载失败: %w", err)
	}
	return ch, nil
}

// InstallHelmFromRepo 仓库模式安装/升级（repoId + chart 名 + 版本，免 tgz 上传）
func (s *K8sClusterService) InstallHelmFromRepo(clusterID uint, repoID uint, namespace, releaseName, chartRef, version, valuesYAML string) (*HelmReleaseInfo, error) {
	if namespace == "" || releaseName == "" || chartRef == "" {
		return nil, newK8sErr(ErrCodeNamespaceReq, "namespace/releaseName/chart 必填")
	}
	var repo model.K8sHelmRepo
	if err := global.GVA_DB.First(&repo, repoID).Error; err != nil {
		return nil, newK8sErr(ErrCodeClusterNotFound, "仓库不存在")
	}
	vals, err := parseValuesYAML(valuesYAML)
	if err != nil {
		return nil, err
	}
	actionCfg, err := helmActionConfig(clusterID, namespace)
	if err != nil {
		return nil, err
	}
	ch, err := locateRepoChart(&repo, chartRef, version)
	if err != nil {
		return nil, err
	}
	return installOrUpgrade(actionCfg, namespace, releaseName, ch, vals)
}

// HelmChartMeta chart 元数据（values 表单模式：schema JSON + 默认值 JSON，均原文回传由前端渲染表单）
type HelmChartMeta struct {
	ChartName    string `json:"chartName"`
	ChartVersion string `json:"chartVersion"`
	AppVersion   string `json:"appVersion"`
	Schema       string `json:"schema"`   // values.schema.json 原文（chart 未提供时为空）
	Defaults     string `json:"defaults"` // 默认 values 的 JSON 序列化（恒为合法 JSON）
}

func chartMetaOf(ch *chart.Chart) *HelmChartMeta {
	m := &HelmChartMeta{Defaults: "{}"}
	if ch.Metadata != nil {
		m.ChartName = ch.Name()
		m.ChartVersion = ch.Metadata.Version
		m.AppVersion = ch.Metadata.AppVersion
	}
	if len(ch.Schema) > 0 {
		m.Schema = string(ch.Schema)
	}
	if ch.Values != nil {
		if b, err := json.Marshal(ch.Values); err == nil {
			m.Defaults = string(b)
		}
	}
	return m
}

// GetHelmChartMetaFromRepo 仓库模式读取 chart 元数据（values 表单数据源）
func (s *K8sClusterService) GetHelmChartMetaFromRepo(repoID uint, chartRef, version string) (*HelmChartMeta, error) {
	if chartRef == "" {
		return nil, newK8sErr(ErrCodeNamespaceReq, "chart 名必填")
	}
	var repo model.K8sHelmRepo
	if err := global.GVA_DB.First(&repo, repoID).Error; err != nil {
		return nil, newK8sErr(ErrCodeClusterNotFound, "仓库不存在")
	}
	ch, err := locateRepoChart(&repo, chartRef, version)
	if err != nil {
		return nil, err
	}
	return chartMetaOf(ch), nil
}

// InspectHelmChartMeta 上传模式读取 chart 元数据（安装前预览 schema，不落集群）
func (s *K8sClusterService) InspectHelmChartMeta(tgz io.Reader) (*HelmChartMeta, error) {
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
	return chartMetaOf(ch), nil
}

// HelmReleaseDetail release 详情（values + 渲染 manifest——manifest 含敏感面，仅 888）
type HelmReleaseDetail struct {
	HelmReleaseInfo
	Values   string `json:"values"`
	Manifest string `json:"manifest"`
}

// GetHelmReleaseDetail release 详情（当前版本 values 与渲染清单）
func (s *K8sClusterService) GetHelmReleaseDetail(clusterID uint, namespace, name string) (*HelmReleaseDetail, error) {
	if namespace == "" || name == "" {
		return nil, newK8sErr(ErrCodeNamespaceReq, "namespace/name 必填")
	}
	actionCfg, err := helmActionConfig(clusterID, namespace)
	if err != nil {
		return nil, err
	}
	rel, err := action.NewGet(actionCfg).Run(name)
	if err != nil {
		return nil, fmt.Errorf("release 获取失败: %w", err)
	}
	detail := &HelmReleaseDetail{HelmReleaseInfo: *releaseInfo(rel), Manifest: rel.Manifest}
	if b, verr := yaml.Marshal(rel.Config); verr == nil {
		detail.Values = string(b)
	}
	return detail, nil
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
