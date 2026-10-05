// Package api 容器实时管理接口
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/container/service"
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

// CreateContainer 创建容器
// @Tags DockerContainer
// @Summary 创建容器（端口/挂载/环境变量/资源限制/重启策略）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body service.CreateContainerReq true "创建参数"
// @Success 200 {object} response.Response{data=object} "创建成功（容器ID）"
// @Router /container/container [post]
func (a *containerApi) CreateContainer(c *gin.Context) {
	var req service.CreateContainerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	id, err := ctSvc.CreateContainer(req.EndpointID, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"id": id}, "创建成功", c)
}

// GetEventList 容器事件列表
// @Tags DockerContainer
// @Summary 容器事件列表（巡检循环区间拉取落库，倒序 200 条）
// @Security ApiKeyAuth
// @Produce application/json
// @Param endpointId query int false "接入点ID（空=全部）"
// @Param action query string false "动作过滤（如 start/die）"
// @Success 200 {object} response.Response{data=[]model.DockerEventLog} "获取成功"
// @Router /container/event/list [get]
func (a *containerApi) GetEventList(c *gin.Context) {
	endpointID, _ := strconv.ParseUint(c.Query("endpointId"), 10, 64)
	list, err := ctSvc.GetEventList(uint(endpointID), c.Query("action"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// ListImages 镜像列表
// @Tags DockerImage
// @Summary 接入点镜像列表（实时）
// @Security ApiKeyAuth
// @Produce application/json
// @Param endpointId query int true "接入点ID"
// @Success 200 {object} response.Response{data=[]service.ImageView} "获取成功"
// @Router /container/image/list [get]
func (a *containerApi) ListImages(c *gin.Context) {
	endpointID, err := strconv.ParseUint(c.Query("endpointId"), 10, 64)
	if err != nil || endpointID == 0 {
		response.FailWithMessage("endpointId 无效", c)
		return
	}
	list, err := ctSvc.ListImages(uint(endpointID))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// PullImage 拉取镜像（异步）
// @Tags DockerImage
// @Summary 拉取镜像（后台任务，状态经 pull-status 查询）
// @Security ApiKeyAuth
// @Produce application/json
// @Param endpointId query int true "接入点ID"
// @Param ref query string true "镜像引用（如 nginx:alpine）"
// @Success 200 {object} response.Response{msg=string} "已开始拉取"
// @Router /container/image/pull [post]
func (a *containerApi) PullImage(c *gin.Context) {
	endpointID, err := strconv.ParseUint(c.Query("endpointId"), 10, 64)
	if err != nil || endpointID == 0 {
		response.FailWithMessage("endpointId 无效", c)
		return
	}
	if err := ctSvc.PullImage(uint(endpointID), c.Query("ref")); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("已开始拉取", c)
}

// PullStatus 拉取状态
// @Tags DockerImage
// @Summary 拉取任务状态查询
// @Security ApiKeyAuth
// @Produce application/json
// @Param endpointId query int true "接入点ID"
// @Param ref query string true "镜像引用"
// @Success 200 {object} response.Response{data=string} "获取成功"
// @Router /container/image/pull-status [get]
func (a *containerApi) PullStatus(c *gin.Context) {
	endpointID, err := strconv.ParseUint(c.Query("endpointId"), 10, 64)
	if err != nil || endpointID == 0 {
		response.FailWithMessage("endpointId 无效", c)
		return
	}
	response.OkWithDetailed(ctSvc.GetPullStatus(uint(endpointID), c.Query("ref")), "获取成功", c)
}

// RemoveImage 删除镜像
// @Tags DockerImage
// @Summary 删除镜像（force）
// @Security ApiKeyAuth
// @Produce application/json
// @Param endpointId query int true "接入点ID"
// @Param ref query string true "镜像引用或ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /container/image [delete]
func (a *containerApi) RemoveImage(c *gin.Context) {
	endpointID, err := strconv.ParseUint(c.Query("endpointId"), 10, 64)
	if err != nil || endpointID == 0 {
		response.FailWithMessage("endpointId 无效", c)
		return
	}
	if err := ctSvc.RemoveImage(uint(endpointID), c.Query("ref")); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
