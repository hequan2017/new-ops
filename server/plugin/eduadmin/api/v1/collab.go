// Package v1 eduadmin API：跨校协同（需求 3.6-27/28）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SaveEduShareResource 保存共享资源
// @Tags Eduadmin
// @Summary 新增跨校共享资源（课程/设备/师资/场地）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduShareResource true "资源"
// @Success 200 {object} response.Response{data=model.EduShareResource} "保存成功"
// @Router /eduadmin/share/resource [post]
func (a *EduadminApi) SaveEduShareResource(c *gin.Context) {
	var in model.EduShareResource
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduShareResource(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetEduShareResources 资源列表
// @Tags Eduadmin
// @Summary 共享资源列表（类别过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param category query string false "课程/设备/师资/场地"
// @Success 200 {object} response.Response{data=[]model.EduShareResource} "获取成功"
// @Router /eduadmin/share/resource/list [get]
func (a *EduadminApi) GetEduShareResources(c *gin.Context) {
	list, err := service.EduadminApp.GetEduShareResources(c.Query("category"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// ScheduleEduShareResource 资源调度
// @Tags Eduadmin
// @Summary 资源调度申请（规则校验：非归属校/可共享/有容量，通过即扣容量）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.scheduleReq true "调度请求"
// @Success 200 {object} response.Response{data=model.EduShareResource} "调度成功"
// @Router /eduadmin/share/schedule [post]
func (a *EduadminApi) ScheduleEduShareResource(c *gin.Context) {
	var in scheduleReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.ScheduleEduShareResource(in.ID, in.School)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "调度成功", c)
}

// scheduleReq 调度请求
type scheduleReq struct {
	ID     uint   `json:"id" binding:"required"`
	School string `json:"school" binding:"required"`
}

// AnalyzeEduJointBudget 联合预算分析
// @Tags Eduadmin
// @Summary 联合预算分析（多校经费聚合+统筹分配建议 下限 10%）
// @Security ApiKeyAuth
// @Produce application/json
// @Param year query int true "年度"
// @Success 200 {object} response.Response{data=service.SchoolBudgetReport} "分析完成"
// @Router /eduadmin/jointBudget [get]
func (a *EduadminApi) AnalyzeEduJointBudget(c *gin.Context) {
	year, _ := strconv.Atoi(c.Query("year"))
	out, err := service.EduadminApp.AnalyzeEduJointBudget(year)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "分析完成", c)
}

// GetEduBestPractices 培训材料/最佳实践
// @Tags Eduadmin
// @Summary 培训材料与最佳实践条目列表（类别过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param category query string false "培训材料/最佳实践"
// @Success 200 {object} response.Response{data=[]model.EduBestPractice} "获取成功"
// @Router /eduadmin/share/practices [get]
func (a *EduadminApi) GetEduBestPractices(c *gin.Context) {
	list, err := service.EduadminApp.GetEduBestPractices(c.Query("category"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
