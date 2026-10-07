// Package api 白泽流水线接口层
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
	plModel "github.com/hequan2017/new-ops/server/plugin/pipeline/model"
	"github.com/hequan2017/new-ops/server/plugin/pipeline/service"
)

var PipelineApi = new(pipelineApi)

// pipelineSvc 指针接收者服务实例（与 asset/job 插件出口惯例一致）
var pipelineSvc = new(service.PipelineService)

// buildSvc 构建服务实例（build.go 共用）
var buildSvc = new(service.PipelineBuildService)

type pipelineApi struct{}

// CreatePipeline 创建流水线
// @Tags Pipeline
// @Summary 创建流水线（含阶段与步骤）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body plModel.Pipeline true "名称/描述/阶段[步骤]"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /pipeline [post]
func (p *pipelineApi) CreatePipeline(c *gin.Context) {
	var pl plModel.Pipeline
	if err := c.ShouldBindJSON(&pl); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := pipelineSvc.CreatePipeline(&pl); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"ID": pl.ID}, "创建成功", c)
}

// UpdatePipeline 更新流水线（阶段/步骤整体替换）
// @Tags Pipeline
// @Summary 更新流水线
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body plModel.Pipeline true "含ID与完整嵌套定义"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /pipeline [put]
func (p *pipelineApi) UpdatePipeline(c *gin.Context) {
	var pl plModel.Pipeline
	if err := c.ShouldBindJSON(&pl); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := pipelineSvc.UpdatePipeline(&pl); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeletePipeline 删除流水线
// @Tags Pipeline
// @Summary 删除流水线（级联清阶段/步骤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "流水线ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /pipeline [delete]
func (p *pipelineApi) DeletePipeline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := pipelineSvc.DeletePipeline(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetPipelineList 流水线列表
// @Tags Pipeline
// @Summary 流水线列表（keyword 过滤，含嵌套定义）
// @Security ApiKeyAuth
// @Produce application/json
// @Param keyword query string false "名称/描述模糊"
// @Success 200 {object} response.Response{data=[]plModel.Pipeline} "获取成功"
// @Router /pipeline/list [get]
func (p *pipelineApi) GetPipelineList(c *gin.Context) {
	list, err := pipelineSvc.GetPipelineList(c.Query("keyword"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// GetPipelineDetail 流水线详情
// @Tags Pipeline
// @Summary 流水线详情（阶段/步骤按 Sort 排序）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "流水线ID"
// @Success 200 {object} response.Response{data=plModel.Pipeline} "获取成功"
// @Router /pipeline/find [get]
func (p *pipelineApi) GetPipelineDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	pl, err := pipelineSvc.GetPipelineDetail(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(pl, "获取成功", c)
}
