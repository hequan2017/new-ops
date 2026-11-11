// Package v1 eduadmin API：反馈闭环（需求 3.6-36）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// reviewReq 评价请求
type reviewReq struct {
	TicketId uint   `json:"ticketId" binding:"required"`
	Score    string `json:"score" binding:"required"`
	Comment  string `json:"comment"`
}

// ReviewEduTicket 评价
// @Tags Eduadmin
// @Summary 工单满意度评价（办结态；差评自动回流办理+时间线留痕）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.reviewReq true "评价请求"
// @Success 200 {object} response.Response{data=model.EduTicketReview} "评价成功"
// @Router /eduadmin/review [post]
func (a *EduadminApi) ReviewEduTicket(c *gin.Context) {
	var in reviewReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.ReviewEduTicket(in.TicketId, in.Score, in.Comment)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "评价成功", c)
}

// ArchiveEduTicket 归档
// @Tags Eduadmin
// @Summary 工单归档（评价态→归档终态）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "工单ID"
// @Success 200 {object} response.Response{data=model.EduTicket} "归档成功"
// @Router /eduadmin/review/archive [post]
func (a *EduadminApi) ArchiveEduTicket(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	out, err := service.EduadminApp.ArchiveEduTicket(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "归档成功", c)
}

// GetEduTicketReviews 评价列表
// @Tags Eduadmin
// @Summary 工单评价列表（评分过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param score query string false "好评/差评"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/review/list [get]
func (a *EduadminApi) GetEduTicketReviews(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.EduadminApp.GetEduTicketReviews(c.Query("score"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}
