// Package v1 eduadmin API：能耗优化（需求 3.6-20）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SaveEduEnergyRecord 新增能耗记录
// @Tags Eduadmin
// @Summary 新增楼栋能耗记录（年月/时段）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduEnergyRecord true "能耗记录"
// @Success 200 {object} response.Response{data=model.EduEnergyRecord} "保存成功"
// @Router /eduadmin/energy [post]
func (a *EduadminApi) SaveEduEnergyRecord(c *gin.Context) {
	var in model.EduEnergyRecord
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduEnergyRecord(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetEduEnergyRecords 能耗记录列表
// @Tags Eduadmin
// @Summary 能耗记录列表（楼栋过滤，年月倒序）
// @Security ApiKeyAuth
// @Produce application/json
// @Param building query string false "楼栋"
// @Success 200 {object} response.Response{data=[]model.EduEnergyRecord} "获取成功"
// @Router /eduadmin/energy/list [get]
func (a *EduadminApi) GetEduEnergyRecords(c *gin.Context) {
	list, err := service.EduadminApp.GetEduEnergyRecords(c.Query("building"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// AnalyzeEduEnergy 能耗分析
// @Tags Eduadmin
// @Summary 能耗聚合分析（楼栋×时段+同比异常>20%+节能建议模板）
// @Security ApiKeyAuth
// @Produce application/json
// @Param year query int true "年度"
// @Param month query int true "月份"
// @Success 200 {object} response.Response{data=map[string]any} "分析完成"
// @Router /eduadmin/energy/analyze [get]
func (a *EduadminApi) AnalyzeEduEnergy(c *gin.Context) {
	year, _ := strconv.Atoi(c.Query("year"))
	month, _ := strconv.Atoi(c.Query("month"))
	usages, suggestions, err := service.EduadminApp.AnalyzeEduEnergy(year, month)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"usages": usages, "suggestions": suggestions}, "分析完成", c)
}
