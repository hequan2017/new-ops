// Package v1 eduadmin API：智能分派（需求 3.6-36）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SaveEduDispatchRule 保存规则
// @Tags Eduadmin
// @Summary 新建/更新分派规则（关键词→部门，优先级大者优先）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduDispatchRule true "规则"
// @Success 200 {object} response.Response{data=model.EduDispatchRule} "保存成功"
// @Router /eduadmin/dispatch/rule [post]
func (a *EduadminApi) SaveEduDispatchRule(c *gin.Context) {
	var in model.EduDispatchRule
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduDispatchRule(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// AutoDispatchEduTicket 自动分派
// @Tags Eduadmin
// @Summary 自动分派（关键词规则命中→推进分派；零命中留待派池）
// @Security ApiKeyAuth
// @Produce application/json
// @Param ticketId query int true "工单ID"
// @Success 200 {object} response.Response{data=model.EduDispatchLog} "分派完成"
// @Router /eduadmin/dispatch/auto [post]
func (a *EduadminApi) AutoDispatchEduTicket(c *gin.Context) {
	ticketId, _ := strconv.ParseUint(c.Query("ticketId"), 10, 64)
	out, err := service.EduadminApp.AutoDispatchEduTicket(uint(ticketId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "分派完成", c)
}

// ManualDispatchEduTicket 人工兜底
// @Tags Eduadmin
// @Summary 人工兜底分派（待派池→指定部门）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.manualDispatchReq true "分派请求"
// @Success 200 {object} response.Response{data=model.EduDispatchLog} "分派完成"
// @Router /eduadmin/dispatch/manual [post]
func (a *EduadminApi) ManualDispatchEduTicket(c *gin.Context) {
	var in manualDispatchReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.ManualDispatchEduTicket(in.TicketId, in.Dept)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "分派完成", c)
}

// manualDispatchReq 人工分派请求
type manualDispatchReq struct {
	TicketId uint   `json:"ticketId" binding:"required"`
	Dept     string `json:"dept" binding:"required"`
}

// GetEduPendingTickets 待派池
// @Tags Eduadmin
// @Summary 待派池（受理态工单，升序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/dispatch/pending [get]
func (a *EduadminApi) GetEduPendingTickets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.EduadminApp.GetEduPendingTickets(page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetEduDispatchLogs 分派记录
// @Tags Eduadmin
// @Summary 分派记录（工单过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param ticketId query int false "工单ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/dispatch/logs [get]
func (a *EduadminApi) GetEduDispatchLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	ticketId, _ := strconv.ParseUint(c.Query("ticketId"), 10, 64)
	list, total, err := service.EduadminApp.GetEduDispatchLogs(uint(ticketId), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}
