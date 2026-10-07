// Package api 三级 RBAC 之命名空间授权接口（授权管理仅 888；ns 可见性 888+9528）
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/k8s/model"
	"github.com/hequan2017/new-ops/server/utils"
)

type k8sGrant struct{}

// nsGrantReq 授权请求体
type nsGrantReq struct {
	ClusterID uint   `json:"clusterId" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	UserID    uint   `json:"userId" binding:"required"`
}

// ListNsVisibility 当前用户视角的命名空间可见性（超管全量 granted=true）
// @Tags K8sGrant
// @Summary 命名空间可见性列表（普通用户仅见授权项）
// @Security ApiKeyAuth
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Success 200 {object} response.Response{data=[]service.NsVisibilityItem} "获取成功"
// @Router /k8s/cluster/ns-visibility [get]
func (g *k8sGrant) ListNsVisibility(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	list, err := k8sClusterService.ListNamespacesForUser(clusterID, utils.GetUserID(c), utils.GetUserAuthorityId(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// CreateNsGrant 授予命名空间可见性
// @Tags K8sGrant
// @Summary 授予用户命名空间可见性（写操作仅 888）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param body body nsGrantReq true "授权信息"
// @Success 200 {object} response.Response{msg=string} "授权成功"
// @Router /k8s/grant [post]
func (g *k8sGrant) CreateNsGrant(c *gin.Context) {
	var req nsGrantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数无效: "+err.Error(), c)
		return
	}
	gr := &model.K8sNsGrant{ClusterID: req.ClusterID, Namespace: req.Namespace, UserID: req.UserID}
	if err := k8sClusterService.CreateNsGrant(gr); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("授权成功", c)
}

// DeleteNsGrant 收回授权
// @Tags K8sGrant
// @Summary 收回命名空间授权（写操作仅 888）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "授权ID"
// @Success 200 {object} response.Response{msg=string} "收回成功"
// @Router /k8s/grant [delete]
func (g *k8sGrant) DeleteNsGrant(c *gin.Context) {
	id, ok := parseUintQ(c, "id")
	if !ok {
		return
	}
	if err := k8sClusterService.DeleteNsGrant(id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("收回成功", c)
}

// ListNsGrants 集群授权清单
// @Tags K8sGrant
// @Summary 集群命名空间授权清单（含用户名；仅 888）
// @Security ApiKeyAuth
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Success 200 {object} response.Response{data=[]service.NsGrantItem} "获取成功"
// @Router /k8s/grant/list [get]
func (g *k8sGrant) ListNsGrants(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	list, err := k8sClusterService.ListNsGrants(clusterID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}
