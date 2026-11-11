// Package v1 eduadmin API：AI-12345 工单（需求 3.6-36）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// CreateEduTicket 创建工单
// @Tags Eduadmin
// @Summary 创建 AI-12345 工单（多渠道受理+时限类别+唯一编号）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduTicket true "工单"
// @Success 200 {object} response.Response{data=model.EduTicket} "受理成功"
// @Router /eduadmin/ticket [post]
func (a *EduadminApi) CreateEduTicket(c *gin.Context) {
	var in model.EduTicket
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.CreateEduTicket(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "受理成功", c)
}

// GetEduTickets 工单列表
// @Tags Eduadmin
// @Summary 工单列表（渠道/状态过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param channel query string false "网页/企微/小程序"
// @Param status query string false "受理/分派/办理/办结/评价/归档"
// @Param demandType query string false "诉求类型关键字"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/ticket/list [get]
func (a *EduadminApi) GetEduTickets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.EduadminApp.GetEduTickets(c.Query("channel"), c.Query("status"), c.Query("demandType"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetEduTicket 工单详情
// @Tags Eduadmin
// @Summary 工单详情（按 ID）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "工单ID"
// @Success 200 {object} response.Response{data=model.EduTicket} "获取成功"
// @Router /eduadmin/ticket [get]
func (a *EduadminApi) GetEduTicket(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	out, err := service.EduadminApp.GetEduTicket(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "获取成功", c)
}

// AdvanceEduTicket 状态推进
// @Tags Eduadmin
// @Summary 工单状态推进（受理→分派→办理→办结→评价→归档 线性下一节点）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "工单ID"
// @Success 200 {object} response.Response{data=model.EduTicket} "推进成功"
// @Router /eduadmin/ticket/advance [post]
func (a *EduadminApi) AdvanceEduTicket(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	out, err := service.EduadminApp.AdvanceEduTicket(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "推进成功", c)
}
