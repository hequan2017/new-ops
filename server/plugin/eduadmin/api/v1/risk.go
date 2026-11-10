// Package v1 eduadmin API：财务风险预警（需求 3.6-17）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// ScanEduRisk 风险扫描
// @Tags Eduadmin
// @Summary 财务风险扫描（支出 3σ 异常+现金流阈值，幂等落库）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body service.riskScanReq true "扫描请求"
// @Success 200 {object} response.Response{data=[]model.EduRiskAlert} "扫描完成"
// @Router /eduadmin/risk/scan [post]
func (a *EduadminApi) ScanEduRisk(c *gin.Context) {
	var in service.RiskScanReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.ScanEduRisk(in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "扫描完成", c)
}

// GetEduRiskAlerts 预警列表
// @Tags Eduadmin
// @Summary 财务预警列表（类型/处理状态过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param type query string false "支出异常/现金流不足"
// @Param handled query string false "yes/no"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/risk/list [get]
func (a *EduadminApi) GetEduRiskAlerts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.EduadminApp.GetEduRiskAlerts(c.Query("type"), c.Query("handled"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// HandleEduRiskAlert 标记处理
// @Tags Eduadmin
// @Summary 预警标记已处理
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "预警ID"
// @Success 200 {object} response.Response "处理成功"
// @Router /eduadmin/risk/handle [post]
func (a *EduadminApi) HandleEduRiskAlert(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if err := service.EduadminApp.HandleEduRiskAlert(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("处理成功", c)
}
