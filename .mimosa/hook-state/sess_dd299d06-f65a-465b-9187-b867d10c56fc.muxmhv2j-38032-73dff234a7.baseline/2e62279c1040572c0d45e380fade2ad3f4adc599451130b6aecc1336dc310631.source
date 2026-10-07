// Package api 网络与卷与统计接口
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/container/service"
)

// ListNetworks 网络列表
// @Tags DockerNetwork
// @Summary 接入点网络列表（实时）
// @Security ApiKeyAuth
// @Produce application/json
// @Param endpointId query int true "接入点ID"
// @Success 200 {object} response.Response{data=[]service.NetworkView} "获取成功"
// @Router /container/network/list [get]
func (a *containerApi) ListNetworks(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("endpointId"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("endpointId 无效", c)
		return
	}
	list, err := ctSvc.ListNetworks(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// CreateNetwork 创建网络
// @Tags DockerNetwork
// @Summary 创建自定义网络（bridge/子网）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body service.CreateNetworkReq true "名称/驱动/子网"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /container/network [post]
func (a *containerApi) CreateNetwork(c *gin.Context) {
	var req service.CreateNetworkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := ctSvc.CreateNetwork(req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// RemoveNetwork 删除网络
// @Tags DockerNetwork
// @Summary 删除自定义网络（内置拒绝）
// @Security ApiKeyAuth
// @Produce application/json
// @Param endpointId query int true "接入点ID"
// @Param name query string true "网络名"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /container/network [delete]
func (a *containerApi) RemoveNetwork(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("endpointId"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("endpointId 无效", c)
		return
	}
	if err := ctSvc.RemoveNetwork(uint(id), c.Query("name")); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ListVolumes 卷列表
// @Tags DockerVolume
// @Summary 接入点卷列表（实时）
// @Security ApiKeyAuth
// @Produce application/json
// @Param endpointId query int true "接入点ID"
// @Success 200 {object} response.Response{data=[]service.VolumeView} "获取成功"
// @Router /container/volume/list [get]
func (a *containerApi) ListVolumes(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("endpointId"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("endpointId 无效", c)
		return
	}
	list, err := ctSvc.ListVolumes(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// RemoveVolume 删除卷
// @Tags DockerVolume
// @Summary 删除卷（占用时 daemon 拒绝）
// @Security ApiKeyAuth
// @Produce application/json
// @Param endpointId query int true "接入点ID"
// @Param name query string true "卷名"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /container/volume [delete]
func (a *containerApi) RemoveVolume(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("endpointId"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("endpointId 无效", c)
		return
	}
	if err := ctSvc.RemoveVolume(uint(id), c.Query("name")); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ContainerStats 容器即时统计
// @Tags DockerContainer
// @Summary 容器资源统计（CPU/内存/网络/块IO 单次采样）
// @Security ApiKeyAuth
// @Produce application/json
// @Param endpointId query int true "接入点ID"
// @Param id query string true "容器ID"
// @Success 200 {object} response.Response{data=service.StatsView} "获取成功"
// @Router /container/container/stats [get]
func (a *containerApi) ContainerStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("endpointId"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("endpointId 无效", c)
		return
	}
	cid := c.Query("id")
	if cid == "" {
		response.FailWithMessage("容器 ID 不能为空", c)
		return
	}
	v, err := ctSvc.ContainerStats(uint(id), cid)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(v, "获取成功", c)
}
