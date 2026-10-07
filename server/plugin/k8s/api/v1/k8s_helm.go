// Package api M8 K3 起步：Helm release 管理接口（list/history 只读；install/uninstall/rollback 写级仅 888）
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
)

type k8sHelm struct{}

// ListHelmReleases Helm release 列表
// @Tags K8sHelm
// @Summary Helm release 列表（namespace 空=全部）
// @Security ApiKeyAuth
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param namespace query string false "命名空间（空=全部）"
// @Success 200 {object} response.Response{data=[]service.HelmReleaseInfo} "获取成功"
// @Router /k8s/helm/list [get]
func (h *k8sHelm) ListHelmReleases(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	list, err := k8sClusterService.ListHelmReleases(clusterID, c.Query("namespace"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// GetHelmHistory release 历史
// @Tags K8sHelm
// @Summary Helm release 历史版本
// @Security ApiKeyAuth
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "release 名"
// @Success 200 {object} response.Response{data=[]service.HelmHistoryItem} "获取成功"
// @Router /k8s/helm/history [get]
func (h *k8sHelm) GetHelmHistory(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	ns, name := c.Query("namespace"), c.Query("name")
	if ns == "" || name == "" {
		response.FailWithMessage("namespace/name 必填", c)
		return
	}
	list, err := k8sClusterService.GetHelmHistory(clusterID, ns, name)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// InstallHelmRelease 安装 release（chart tgz 上传 + values YAML 覆盖）
// @Tags K8sHelm
// @Summary 安装 Helm release（写操作仅 888）
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param namespace formData string true "命名空间"
// @Param releaseName formData string true "release 名"
// @Param values formData string false "values YAML（可选）"
// @Param chart formData file true "chart .tgz 包"
// @Success 200 {object} response.Response{data=service.HelmReleaseInfo,msg=string} "安装成功"
// @Router /k8s/helm/install [post]
func (h *k8sHelm) InstallHelmRelease(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	ns := c.PostForm("namespace")
	releaseName := c.PostForm("releaseName")
	if ns == "" || releaseName == "" {
		response.FailWithMessage("namespace/releaseName 必填", c)
		return
	}
	fh, err := c.FormFile("chart")
	if err != nil {
		response.FailWithMessage("chart 包必填（.tgz）", c)
		return
	}
	f, err := fh.Open()
	if err != nil {
		response.FailWithMessage("chart 包读取失败: "+err.Error(), c)
		return
	}
	defer f.Close()
	info, err := k8sClusterService.InstallHelmRelease(clusterID, ns, releaseName, f, c.PostForm("values"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(info, "安装成功", c)
}

// UninstallHelmRelease 卸载 release
// @Tags K8sHelm
// @Summary 卸载 Helm release（写操作仅 888）
// @Security ApiKeyAuth
// @Accept application/x-www-form-urlencoded
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param namespace formData string true "命名空间"
// @Param name formData string true "release 名"
// @Success 200 {object} response.Response{msg=string} "卸载成功"
// @Router /k8s/helm/uninstall [post]
func (h *k8sHelm) UninstallHelmRelease(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	ns, name := c.PostForm("namespace"), c.PostForm("name")
	if ns == "" || name == "" {
		response.FailWithMessage("namespace/name 必填", c)
		return
	}
	if err := k8sClusterService.UninstallHelmRelease(clusterID, ns, name); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("卸载成功", c)
}

// RollbackHelmRelease 回滚 release
// @Tags K8sHelm
// @Summary 回滚 Helm release（revision<=0 回滚到上一版；写操作仅 888）
// @Security ApiKeyAuth
// @Accept application/x-www-form-urlencoded
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param namespace formData string true "命名空间"
// @Param name formData string true "release 名"
// @Param revision formData int false "目标版本（空=上一版）"
// @Success 200 {object} response.Response{msg=string} "回滚成功"
// @Router /k8s/helm/rollback [post]
func (h *k8sHelm) RollbackHelmRelease(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	ns, name := c.PostForm("namespace"), c.PostForm("name")
	if ns == "" || name == "" {
		response.FailWithMessage("namespace/name 必填", c)
		return
	}
	revision, _ := strconv.Atoi(c.PostForm("revision"))
	if err := k8sClusterService.RollbackHelmRelease(clusterID, ns, name, revision); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("回滚成功", c)
}
