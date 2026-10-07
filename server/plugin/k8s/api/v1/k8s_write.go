// Package api K2 写操作接口（仅 888，写操作走 GVA 操作日志中间件）
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
)

type k8sWrite struct{}

// parseWriteParams 写操作参数解析（clusterId/namespace/name 公共）
func parseWriteParams(c *gin.Context) (clusterID uint, namespace, name string, ok bool) {
	cid, err := strconv.ParseUint(c.Query("clusterId"), 10, 64)
	if err != nil || cid == 0 {
		response.FailWithMessage("clusterId 无效", c)
		return 0, "", "", false
	}
	ns := c.PostForm("namespace")
	name = c.PostForm("name")
	if ns == "" || name == "" {
		response.FailWithMessage("namespace/name 必填", c)
		return 0, "", "", false
	}
	return uint(cid), ns, name, true
}

// ScaleDeployment Deployment 扩缩容
// @Tags K8sWrite
// @Summary Deployment 扩缩容
// @Security ApiKeyAuth
// @Accept application/x-www-form-urlencoded
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param namespace formData string true "命名空间"
// @Param name formData string true "Deployment 名"
// @Param replicas formData int true "目标副本数（0-500）"
// @Success 200 {object} response.Response{msg=string} "扩缩容成功"
// @Router /k8s/deployment/scale [post]
func (w *k8sWrite) ScaleDeployment(c *gin.Context) {
	clusterID, err := strconv.ParseUint(c.Query("clusterId"), 10, 64)
	if err != nil || clusterID == 0 {
		response.FailWithMessage("clusterId 无效", c)
		return
	}
	ns := c.PostForm("namespace")
	name := c.PostForm("name")
	if ns == "" || name == "" {
		response.FailWithMessage("namespace/name 必填", c)
		return
	}
	replicas, err := strconv.ParseInt(c.PostForm("replicas"), 10, 32)
	if err != nil {
		response.FailWithMessage("replicas 无效", c)
		return
	}
	if err := k8sClusterService.ScaleDeployment(uint(clusterID), ns, name, int32(replicas)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("扩缩容已提交", c)
}

// RestartDeployment Deployment 滚动重启
// @Tags K8sWrite
// @Summary Deployment 滚动重启
// @Security ApiKeyAuth
// @Accept application/x-www-form-urlencoded
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param namespace formData string true "命名空间"
// @Param name formData string true "Deployment 名"
// @Success 200 {object} response.Response{msg=string} "滚动重启已提交"
// @Router /k8s/deployment/restart [post]
func (w *k8sWrite) RestartDeployment(c *gin.Context) {
	clusterID, err := strconv.ParseUint(c.Query("clusterId"), 10, 64)
	if err != nil || clusterID == 0 {
		response.FailWithMessage("clusterId 无效", c)
		return
	}
	ns := c.PostForm("namespace")
	name := c.PostForm("name")
	if ns == "" || name == "" {
		response.FailWithMessage("namespace/name 必填", c)
		return
	}
	if err := k8sClusterService.RestartDeployment(uint(clusterID), ns, name); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("滚动重启已提交", c)
}
