// Package api 容器实时管理接口
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
)

// ListContainers 容器列表
// @Tags DockerContainer
// @Summary 接入点容器列表（实时查询 Docker API）
// @Security ApiKeyAuth
// @Produce application/json
// @Param endpointId query int true "接入点ID"
// @Param all query bool false "含已停止（默认 true）"
// @Success 200 {object} response.Response{data=[]service.ContainerView} "获取成功"
// @Router /container/container/list [get]
func (a *containerApi) ListContainers(c *gin.Context) {
	endpointID, err := strconv.ParseUint(c.Query("endpointId"), 10, 64)
	if err != nil || endpointID == 0 {
		response.FailWithMessage("endpointId 无效", c)
		return
	}
	all := c.DefaultQuery("all", "true") == "true"
	list, err := ctSvc.ListContainers(uint(endpointID), all)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// ContainerAction 容器生命周期动作
// @Tags DockerContainer
// @Summary 容器动作（start/stop/restart/remove）
// @Security ApiKeyAuth
// @Produce application/json
// @Param endpointId query int true "接入点ID"
// @Param id query string true "容器ID"
// @Param action query string true "动作"
// @Param force query bool false "强制删除"
// @Success 200 {object} response.Response "操作成功"
// @Router /container/container/action [post]
func (a *containerApi) ContainerAction(c *gin.Context) {
	endpointID, err := strconv.ParseUint(c.Query("endpointId"), 10, 64)
	if err != nil || endpointID == 0 {
		response.FailWithMessage("endpointId 无效", c)
		return
	}
	cid := c.Query("id")
	if cid == "" {
		response.FailWithMessage("容器 ID 不能为空", c)
		return
	}
	action := c.Query("action")
	force := c.Query("force") == "true"
	if err := ctSvc.ContainerAction(uint(endpointID), cid, action, force); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}
