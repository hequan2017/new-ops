// Package v1 oplib API：算子库接口层（参数绑定与响应，禁止业务逻辑）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/oplib/model"
	"github.com/hequan2017/new-ops/server/plugin/oplib/service"
)

var Api = new(oplibApi)

type oplibApi struct{}

// GetOperators 算子列表
// @Tags Oplib
// @Summary 算子列表（分类/实现过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param category query string false "数据处理/特征/检索/后处理/评估"
// @Param implType query string false "builtin/script"
// @Success 200 {object} response.Response{data=[]model.OplibOperator} "获取成功"
// @Router /oplib/operator/list [get]
func (a *oplibApi) GetOperators(c *gin.Context) {
	list, err := service.OplibApp.GetOperators(c.Query("category"), c.Query("implType"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveOperator 保存算子
// @Tags Oplib
// @Summary 新建/更新算子（schema 对象即检）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.OplibOperator true "算子"
// @Success 200 {object} response.Response{data=model.OplibOperator} "保存成功"
// @Router /oplib/operator [post]
func (a *oplibApi) SaveOperator(c *gin.Context) {
	var in model.OplibOperator
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	op, err := service.OplibApp.SaveOperator(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(op, "保存成功", c)
}

// DeleteOperator 删除算子
// @Tags Oplib
// @Summary 删除算子（软删除）
// @Security ApiKeyAuth
// @Param id query int true "算子ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /oplib/operator [delete]
func (a *oplibApi) DeleteOperator(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.OplibApp.DeleteOperator(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// RunOperator 执行算子
// @Tags Oplib
// @Summary 执行算子（schema 校验→执行→留痕；builtin）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.RunIn true "执行请求"
// @Success 200 {object} response.Response{data=model.OplibRunLog} "执行完成"
// @Router /oplib/operator/run [post]
func (a *oplibApi) RunOperator(c *gin.Context) {
	var in service.RunIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	runLog, err := service.OplibApp.ExecuteOperator(in.ID, in.Args)
	if err != nil && runLog == nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(runLog, "执行完成", c)
}

// GetRunLogs 执行留痕
// @Tags Oplib
// @Summary 算子执行留痕列表（算子过滤，倒序分页）
// @Security ApiKeyAuth
// @Param operatorId query int false "算子ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /oplib/operator/runLogs [get]
func (a *oplibApi) GetRunLogs(c *gin.Context) {
	operatorId, _ := strconv.ParseUint(c.Query("operatorId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.OplibApp.GetRunLogs(uint(operatorId), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetPipelines 管道列表
// @Tags Oplib
// @Summary 算子管道列表
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]model.OplibPipeline} "获取成功"
// @Router /oplib/pipeline/list [get]
func (a *oplibApi) GetPipelines(c *gin.Context) {
	list, err := service.OplibApp.GetPipelines()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SavePipeline 保存管道
// @Tags Oplib
// @Summary 新建/更新算子管道（定义拓扑校验即检）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.OplibPipeline true "管道"
// @Success 200 {object} response.Response{data=model.OplibPipeline} "保存成功"
// @Router /oplib/pipeline [post]
func (a *oplibApi) SavePipeline(c *gin.Context) {
	var in model.OplibPipeline
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	p, err := service.OplibApp.SavePipeline(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(p, "保存成功", c)
}

// RunPipeline 执行管道
// @Tags Oplib
// @Summary 执行算子管道（拓扑序执行，整体留痕）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.PipeRunIn true "执行请求"
// @Success 200 {object} response.Response{data=service.PipeResult} "执行完成"
// @Router /oplib/pipeline/run [post]
func (a *oplibApi) RunPipeline(c *gin.Context) {
	var in service.PipeRunIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	res, err := service.OplibApp.RunPipeline(in.ID, in.Input)
	if err != nil && res == nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(res, "执行完成", c)
}

// DeletePipeline 删除管道
// @Tags Oplib
// @Summary 删除算子管道（软删除）
// @Security ApiKeyAuth
// @Param id query int true "管道ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /oplib/pipeline [delete]
func (a *oplibApi) DeletePipeline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.OplibApp.DeletePipeline(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
