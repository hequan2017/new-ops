// Package api 白泽 GPU 算力接口
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
	gpuModel "github.com/hequan2017/new-ops/server/plugin/gpu/model"
	"github.com/hequan2017/new-ops/server/plugin/gpu/service"
	"github.com/hequan2017/new-ops/server/utils"
)

var GpuApi = new(gpuApi)

var gpuSvc = new(service.GpuService)

type gpuApi struct{}

// CreateNode 注册算力节点
// @Tags GpuNode
// @Summary 注册算力节点
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body gpuModel.GpuNode true "名称/显卡型号/卡数/CPU/内存"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /gpu/node [post]
func (a *gpuApi) CreateNode(c *gin.Context) {
	var n gpuModel.GpuNode
	if err := c.ShouldBindJSON(&n); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := gpuSvc.CreateNode(&n); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateNode 更新节点
// @Tags GpuNode
// @Summary 更新节点（总量不可低于运行中占用）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body gpuModel.GpuNode true "含ID"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /gpu/node [put]
func (a *gpuApi) UpdateNode(c *gin.Context) {
	var n gpuModel.GpuNode
	if err := c.ShouldBindJSON(&n); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := gpuSvc.UpdateNode(&n); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteNode 删除节点
// @Tags GpuNode
// @Summary 删除节点（有运行中实例拒绝）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "节点ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /gpu/node [delete]
func (a *gpuApi) DeleteNode(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := gpuSvc.DeleteNode(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ListNode 节点列表（含用量余量）
// @Tags GpuNode
// @Summary 节点列表（含 GPU/CPU/内存 已用余量）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]service.NodeWithUsage} "获取成功"
// @Router /gpu/node/list [get]
func (a *gpuApi) ListNode(c *gin.Context) {
	list, err := gpuSvc.GetNodeList()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// CreateSpec 创建规格
// @Tags GpuSpec
// @Summary 创建产品规格（卡数/CPU/内存/定价）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body gpuModel.GpuSpec true "规格"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /gpu/spec [post]
func (a *gpuApi) CreateSpec(c *gin.Context) {
	var sp gpuModel.GpuSpec
	if err := c.ShouldBindJSON(&sp); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := gpuSvc.CreateSpec(&sp); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteSpec 删除规格
// @Tags GpuSpec
// @Summary 删除规格
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "规格ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /gpu/spec [delete]
func (a *gpuApi) DeleteSpec(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := gpuSvc.DeleteSpec(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ListSpec 规格列表
// @Tags GpuSpec
// @Summary 规格列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]gpuModel.GpuSpec} "获取成功"
// @Router /gpu/spec/list [get]
func (a *gpuApi) ListSpec(c *gin.Context) {
	list, err := gpuSvc.GetSpecList()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// StartInstance 开通实例
// @Tags GpuInstance
// @Summary 开通 GPU 实例（事务内行锁防超卖）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "nodeId/specId/name"
// @Success 200 {object} response.Response{data=gpuModel.GpuInstance} "开通成功"
// @Router /gpu/instance [post]
func (a *gpuApi) StartInstance(c *gin.Context) {
	var req struct {
		NodeID uint   `json:"nodeId" binding:"required"`
		SpecID uint   `json:"specId" binding:"required"`
		Name   string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	inst, err := gpuSvc.StartInstance(req.NodeID, req.SpecID, req.Name,
		utils.GetUserName(c), utils.GetUserID(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(inst, "开通成功", c)
}

// ReleaseInstance 销毁实例
// @Tags GpuInstance
// @Summary 销毁实例（释放配额）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "实例ID"
// @Success 200 {object} response.Response{msg=string} "销毁成功"
// @Router /gpu/instance/release [post]
func (a *gpuApi) ReleaseInstance(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := gpuSvc.ReleaseInstance(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("销毁成功", c)
}

// ListInstance 实例列表
// @Tags GpuInstance
// @Summary 实例列表（status 过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param status query string false "状态"
// @Success 200 {object} response.Response{data=[]gpuModel.GpuInstance} "获取成功"
// @Router /gpu/instance/list [get]
func (a *gpuApi) ListInstance(c *gin.Context) {
	list, err := gpuSvc.GetInstanceList(c.Query("status"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
