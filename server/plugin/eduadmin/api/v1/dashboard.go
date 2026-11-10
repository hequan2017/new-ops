// Package v1 eduadmin API：数据驾驶舱（需求 3.6-24）
package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// GetEduDashboard 驾驶舱
// @Tags Eduadmin
// @Summary 数据驾驶舱（四域计数+来源分布+近 7 日趋势；month 按月下钻）
// @Security ApiKeyAuth
// @Produce application/json
// @Param domain query string false "人事/财务/资产/后勤"
// @Param source query string false "来源关键字"
// @Param month query string false "yyyy-MM 按月下钻"
// @Success 200 {object} response.Response{data=service.DashboardStats} "获取成功"
// @Router /eduadmin/dashboard [get]
func (a *EduadminApi) GetEduDashboard(c *gin.Context) {
	out, err := service.EduadminApp.GetEduDashboard(c.Query("domain"), c.Query("source"), c.Query("month"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "获取成功", c)
}
