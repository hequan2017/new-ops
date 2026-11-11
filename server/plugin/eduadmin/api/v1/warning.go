// Package v1 eduadmin API：办学预警（需求 3.6-25）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SaveEduRunIndex 保存指标
// @Tags Eduadmin
// @Summary 新建/更新办学指标（招生/就业/经费，按名+年幂等）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduRunIndex true "指标"
// @Success 200 {object} response.Response{data=model.EduRunIndex} "保存成功"
// @Router /eduadmin/run/index [post]
func (a *EduadminApi) SaveEduRunIndex(c *gin.Context) {
	var in model.EduRunIndex
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduRunIndex(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetEduRunIndexes 指标列表
// @Tags Eduadmin
// @Summary 办学指标列表（年度过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param year query int false "年度"
// @Success 200 {object} response.Response{data=[]model.EduRunIndex} "获取成功"
// @Router /eduadmin/run/index/list [get]
func (a *EduadminApi) GetEduRunIndexes(c *gin.Context) {
	year, _ := strconv.Atoi(c.Query("year"))
	list, err := service.EduadminApp.GetEduRunIndexes(year)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// ScanEduRunWarning 预警扫描
// @Tags Eduadmin
// @Summary 办学预警扫描（阈值双向判定+幂等落库+招生策略优化报告）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]any} "扫描完成"
// @Router /eduadmin/run/scan [post]
func (a *EduadminApi) ScanEduRunWarning(c *gin.Context) {
	alerts, report, err := service.EduadminApp.ScanEduRunWarning()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"alerts": alerts, "report": report}, "扫描完成", c)
}

// GetEduRunAlerts 预警列表
// @Tags Eduadmin
// @Summary 办学预警列表（年度过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param year query int false "年度"
// @Success 200 {object} response.Response{data=[]model.EduRunAlert} "获取成功"
// @Router /eduadmin/run/alerts [get]
func (a *EduadminApi) GetEduRunAlerts(c *gin.Context) {
	year, _ := strconv.Atoi(c.Query("year"))
	list, err := service.EduadminApp.GetEduRunAlerts(year)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
