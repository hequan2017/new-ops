// Package api 集群接口
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/k8s/model"
)

type k8sCluster struct{}

// CreateCluster 注册集群
// @Tags K8sCluster
// @Summary 注册集群（kubeconfig AES-256-GCM 加密落库 + 连接测试）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "name/remark/kubeconfig"
// @Success 200 {object} response.Response{msg=string} "注册成功"
// @Router /k8s/cluster/create [post]
func (a *k8sCluster) CreateCluster(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		Remark     string `json:"remark"`
		Kubeconfig string `json:"kubeconfig" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c1 := &model.K8sCluster{Name: req.Name, Remark: req.Remark}
	if err := k8sClusterService.CreateCluster(c1, req.Kubeconfig); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("注册成功", c)
}

// DeleteCluster 删除集群
// @Tags K8sCluster
// @Summary 删除集群
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.GetById true "集群ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /k8s/cluster/delete [delete]
func (a *k8sCluster) DeleteCluster(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := k8sClusterService.DeleteCluster(uint(req.ID)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetClusterList 集群列表（kubeconfig 密文不返回）
// @Tags K8sCluster
// @Summary 集群列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param keyword query string false "名称/备注关键字"
// @Success 200 {object} response.Response{data=[]model.K8sCluster} "获取成功"
// @Router /k8s/cluster/list [get]
func (a *k8sCluster) GetClusterList(c *gin.Context) {
	list, err := k8sClusterService.GetClusterList(c.Query("keyword"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// TestCluster 已注册集群连接测试
// @Tags K8sCluster
// @Summary 集群连接测试
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "集群ID"
// @Success 200 {object} response.Response{data=string} "连接成功"
// @Router /k8s/cluster/test [get]
func (a *k8sCluster) TestCluster(c *gin.Context) {
	id, err := strconvUintQuery(c.Query("id"))
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	version, err := k8sClusterService.TestCluster(id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(version, "连接成功", c)
}

// strconvUintQuery 安全解析 uint query 参数
func strconvUintQuery(s string) (uint, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	return uint(n), err
}
