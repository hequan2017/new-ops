// Package v1 eduadmin API：工单进度与督办（需求 3.6-36）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// AddEduTicketTimeline 新增时间线
// @Tags Eduadmin
// @Summary 新增工单时间线（办理单位/人员/进展）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduTicketTimeline true "时间线"
// @Success 200 {object} response.Response{data=model.EduTicketTimeline} "新增成功"
// @Router /eduadmin/progress [post]
func (a *EduadminApi) AddEduTicketTimeline(c *gin.Context) {
	var in model.EduTicketTimeline
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.AddEduTicketTimeline(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "新增成功", c)
}

// GetEduTicketTimelines 时间线列表
// @Tags Eduadmin
// @Summary 工单时间线列表（按工单，倒序）
// @Security ApiKeyAuth
// @Produce application/json
// @Param ticketId query int true "工单ID"
// @Success 200 {object} response.Response{data=[]model.EduTicketTimeline} "获取成功"
// @Router /eduadmin/progress/list [get]
func (a *EduadminApi) GetEduTicketTimelines(c *gin.Context) {
	ticketId, _ := strconv.ParseUint(c.Query("ticketId"), 10, 64)
	list, err := service.EduadminApp.GetEduTicketTimelines(uint(ticketId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// ScanEduTicketOverdue 超时扫描
// @Tags Eduadmin
// @Summary 超时扫描（临近 6h 提醒/超时自动督办，按工单+等级幂等）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.EduTicketUrge} "扫描完成"
// @Router /eduadmin/progress/scan [post]
func (a *EduadminApi) ScanEduTicketOverdue(c *gin.Context) {
	out, err := service.EduadminApp.ScanEduTicketOverdue()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "扫描完成", c)
}

// GetEduTicketUrges 督办记录
// @Tags Eduadmin
// @Summary 督办记录列表（等级过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param level query string false "临近/超时"
// @Success 200 {object} response.Response{data=[]model.EduTicketUrge} "获取成功"
// @Router /eduadmin/progress/urges [get]
func (a *EduadminApi) GetEduTicketUrges(c *gin.Context) {
	list, err := service.EduadminApp.GetEduTicketUrges(c.Query("level"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
