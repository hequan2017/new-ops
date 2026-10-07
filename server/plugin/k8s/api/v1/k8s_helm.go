// Package api M8 K3 起步：Helm release 管理接口（list/history 只读；install/uninstall/rollback 写级仅 888）
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/k8s/model"
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
	list, err := k8sClusterService.ListHelmScoped(clusterID, c.Query("namespace"), scopeUID(c), scopeAuth(c))
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
	if err := k8sClusterService.GuardNamespace(clusterID, scopeUID(c), scopeAuth(c), ns); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := k8sClusterService.GetHelmHistory(clusterID, ns, name)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// InstallHelmRelease 安装/升级 release
// @Tags K8sHelm
// @Summary 安装或升级 Helm release（同名已存在则升级，revision 递增；写操作仅 888）
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param namespace formData string true "命名空间"
// @Param releaseName formData string true "release 名"
// @Param values formData string false "values YAML（可选）"
// @Param chart formData file true "chart .tgz 包"
// @Success 200 {object} response.Response{data=service.HelmReleaseInfo,msg=string} "安装/升级成功"
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

// InstallHelmFromRepo 仓库模式安装/升级
// @Tags K8sHelm
// @Summary 从已登记仓库安装/升级 release（chart 名+版本，免 tgz 上传；写操作仅 888）
// @Security ApiKeyAuth
// @Accept application/x-www-form-urlencoded
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param repoId formData int true "仓库ID"
// @Param namespace formData string true "命名空间"
// @Param releaseName formData string true "release 名"
// @Param chart formData string true "chart 名"
// @Param version formData string false "chart 版本（空=最新）"
// @Param values formData string false "values YAML（可选）"
// @Success 200 {object} response.Response{data=service.HelmReleaseInfo,msg=string} "安装/升级成功"
// @Router /k8s/helm/install-repo [post]
func (h *k8sHelm) InstallHelmFromRepo(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	repoID, err := strconv.ParseUint(c.PostForm("repoId"), 10, 64)
	if err != nil || repoID == 0 {
		response.FailWithMessage("repoId 无效", c)
		return
	}
	info, err := k8sClusterService.InstallHelmFromRepo(clusterID, uint(repoID),
		c.PostForm("namespace"), c.PostForm("releaseName"), c.PostForm("chart"),
		c.PostForm("version"), c.PostForm("values"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(info, "安装/升级成功", c)
}

// GetHelmReleaseDetail release 详情
// @Tags K8sHelm
// @Summary release 详情（当前 values 与渲染 manifest；manifest 含敏感面仅 888）
// @Security ApiKeyAuth
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param name query string true "release 名"
// @Success 200 {object} response.Response{data=service.HelmReleaseDetail} "获取成功"
// @Router /k8s/helm/detail [get]
func (h *k8sHelm) GetHelmReleaseDetail(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	ns, name := c.Query("namespace"), c.Query("name")
	if ns == "" || name == "" {
		response.FailWithMessage("namespace/name 必填", c)
		return
	}
	d, err := k8sClusterService.GetHelmReleaseDetail(clusterID, ns, name)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(d, c)
}

// helmRepoReq 仓库登记请求体
type helmRepoReq struct {
	Name   string `json:"name" binding:"required"`
	URL    string `json:"url" binding:"required"`
	Remark string `json:"remark"`
}

// CreateHelmRepo 登记 chart 仓库
// @Tags K8sHelm
// @Summary 登记 chart 仓库（仅 http/https；写操作仅 888）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param body body helmRepoReq true "仓库信息"
// @Success 200 {object} response.Response{msg=string} "登记成功"
// @Router /k8s/helm/repo [post]
func (h *k8sHelm) CreateHelmRepo(c *gin.Context) {
	var req helmRepoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数无效: "+err.Error(), c)
		return
	}
	r := &model.K8sHelmRepo{Name: req.Name, URL: req.URL, Remark: req.Remark}
	if err := k8sClusterService.CreateHelmRepo(r); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("登记成功", c)
}

// DeleteHelmRepo 删除仓库登记
// @Tags K8sHelm
// @Summary 删除 chart 仓库登记（写操作仅 888）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "仓库ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /k8s/helm/repo [delete]
func (h *k8sHelm) DeleteHelmRepo(c *gin.Context) {
	id, ok := parseUintQ(c, "id")
	if !ok {
		return
	}
	if err := k8sClusterService.DeleteHelmRepo(id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetHelmRepoList 仓库列表
// @Tags K8sHelm
// @Summary chart 仓库列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.K8sHelmRepo} "获取成功"
// @Router /k8s/helm/repo/list [get]
func (h *k8sHelm) GetHelmRepoList(c *gin.Context) {
	list, err := k8sClusterService.GetHelmRepoList()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}
