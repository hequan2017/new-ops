// Package api 白泽工单引擎接口
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/model/common/response"
	wfModel "github.com/hequan2017/new-ops/server/plugin/workflow/model"
	"github.com/hequan2017/new-ops/server/plugin/workflow/service"
	"github.com/hequan2017/new-ops/server/utils"
)

var WorkflowApi = new(workflowApi)

var wfSvc = new(service.WorkflowService)

type workflowApi struct{}

// CreateDefinition 创建工单定义
// @Tags Workflow
// @Summary 创建工单定义（states/transitions JSON + 起始状态）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body wfModel.WfDefinition true "定义"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /workflow/definition [post]
func (a *workflowApi) CreateDefinition(c *gin.Context) {
	var d wfModel.WfDefinition
	if err := c.ShouldBindJSON(&d); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := wfSvc.CreateDefinition(&d); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateDefinition 更新工单定义
// @Tags Workflow
// @Summary 更新工单定义
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body wfModel.WfDefinition true "含ID"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /workflow/definition [put]
func (a *workflowApi) UpdateDefinition(c *gin.Context) {
	var d wfModel.WfDefinition
	if err := c.ShouldBindJSON(&d); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := wfSvc.UpdateDefinition(&d); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteDefinition 删除工单定义
// @Tags Workflow
// @Summary 删除工单定义（有进行中实例拒绝）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "定义ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /workflow/definition [delete]
func (a *workflowApi) DeleteDefinition(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := wfSvc.DeleteDefinition(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ListDefinitions 定义列表
// @Tags Workflow
// @Summary 工单定义列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param keyword query string false "名称模糊"
// @Success 200 {object} response.Response{data=[]wfModel.WfDefinition} "获取成功"
// @Router /workflow/definition/list [get]
func (a *workflowApi) ListDefinitions(c *gin.Context) {
	list, err := wfSvc.GetDefinitionList(c.Query("keyword"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// StartInstance 发起工单
// @Tags Workflow
// @Summary 发起工单实例
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "definitionId/title/bizType/params"
// @Success 200 {object} response.Response{data=wfModel.WfInstance} "发起成功"
// @Router /workflow/instance [post]
func (a *workflowApi) StartInstance(c *gin.Context) {
	var req struct {
		DefinitionID uint   `json:"definitionId" binding:"required"`
		Title        string `json:"title" binding:"required"`
		BizType      string `json:"bizType"`
		Params       string `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	ins, err := wfSvc.StartInstance(req.DefinitionID, req.Title, req.BizType, req.Params,
		utils.GetUserName(c), utils.GetUserID(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(ins, "发起成功", c)
}

// SubmitAction 工单流转
// @Tags Workflow
// @Summary 工单动作（submit/approve/reject/cancel 等，按定义迁移表）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "工单ID"
// @Param action query string true "动作"
// @Param comment query string false "意见"
// @Success 200 {object} response.Response{msg=string} "操作成功"
// @Router /workflow/instance/action [post]
func (a *workflowApi) SubmitAction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	action := c.Query("action")
	if action == "" {
		response.FailWithMessage("action 不能为空", c)
		return
	}
	if _, err := wfSvc.SubmitAction(uint(id), action, c.Query("comment"), utils.GetUserName(c)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

// ListInstances 工单分页
// @Tags Workflow
// @Summary 工单实例列表（普通用户仅自己发起）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.PageInfo true "页码/页大小"
// @Param status query string false "状态"
// @Success 200 {object} response.Response{data=response.PageResult} "获取成功"
// @Router /workflow/instance/list [post]
func (a *workflowApi) ListInstances(c *gin.Context) {
	var info request.PageInfo
	_ = c.ShouldBindJSON(&info)
	list, total, err := wfSvc.GetInstanceList(info.Page, info.PageSize, c.Query("creator"),
		c.Query("status"), utils.GetUserID(c), utils.GetUserAuthorityId(c) == 888)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: info.Page, PageSize: info.PageSize,
	}, "获取成功", c)
}

// GetInstanceLogs 工单流转记录
// @Tags Workflow
// @Summary 工单流转记录（时间线）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "工单ID"
// @Success 200 {object} response.Response{data=[]wfModel.WfActionLog} "获取成功"
// @Router /workflow/instance/logs [get]
func (a *workflowApi) GetInstanceLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	logs, err := wfSvc.GetInstanceLogs(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(logs, "获取成功", c)
}
