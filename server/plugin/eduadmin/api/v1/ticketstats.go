// Package v1 eduadmin API：工单统计报表（需求 3.6-36）
package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// GetEduTicketStats 统计
// @Tags Eduadmin
// @Summary 工单统计（总量/办结率/超时率/满意度/高频诉求 Top5/部门效率）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=service.TicketStats} "获取成功"
// @Router /eduadmin/ticketStats [get]
func (a *EduadminApi) GetEduTicketStats(c *gin.Context) {
	out, err := service.EduadminApp.GetEduTicketStats()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "获取成功", c)
}

// ExportEduTicketStats CSV 导出
// @Tags Eduadmin
// @Summary 工单统计 CSV 导出（复用 I07 口径）
// @Security ApiKeyAuth
// @Produce text/plain
// @Success 200 {object} response.Response{data=string} "导出成功"
// @Router /eduadmin/ticketStats/export [get]
func (a *EduadminApi) ExportEduTicketStats(c *gin.Context) {
	out, err := service.EduadminApp.ExportEduTicketStats()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "导出成功", c)
}
