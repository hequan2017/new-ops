// Package api 白泽容器管理：容器统计历史接口（M8 C2）
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
)

// GetStatsHistory 容器统计历史
// @Tags ContainerStats
// @Summary 容器资源统计历史（采样点，供 ECharts 时间轴）
// @Security ApiKeyAuth
// @Produce application/json
// @Param endpointId query int true "接入点ID"
// @Param id query string true "容器ID"
// @Param hours query int false "窗口小时数（1-168，默认 24）"
// @Success 200 {object} response.Response{data=[]ctModel.DockerStatsSample} "获取成功"
// @Router /container/container/stats/history [get]
func (a *containerApi) GetStatsHistory(c *gin.Context) {
	endpointID, err := strconv.ParseUint(c.Query("endpointId"), 10, 64)
	if err != nil || endpointID == 0 {
		response.FailWithMessage("endpointId 无效", c)
		return
	}
	cid := c.Query("id")
	if cid == "" {
		response.FailWithMessage("容器ID必填", c)
		return
	}
	hours, _ := strconv.Atoi(c.Query("hours"))
	list, err := ctSvc.GetStatsHistory(uint(endpointID), cid, hours)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}
