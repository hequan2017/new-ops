// Package v1 eduadmin API：食堂优化（需求 3.6-21）
package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SaveEduMealRecord 新增消费记录
// @Tags Eduadmin
// @Summary 新增一卡通消费记录（档口/餐段）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduMealRecord true "消费记录"
// @Success 200 {object} response.Response{data=model.EduMealRecord} "保存成功"
// @Router /eduadmin/meal [post]
func (a *EduadminApi) SaveEduMealRecord(c *gin.Context) {
	var in model.EduMealRecord
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduMealRecord(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetEduMealRecords 消费记录列表
// @Tags Eduadmin
// @Summary 一卡通消费记录列表（日期过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param date query string false "日期 yyyy-mm-dd"
// @Success 200 {object} response.Response{data=[]model.EduMealRecord} "获取成功"
// @Router /eduadmin/meal/list [get]
func (a *EduadminApi) GetEduMealRecords(c *gin.Context) {
	list, err := service.EduadminApp.GetEduMealRecords(c.Query("date"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// ForecastEduMeals 用餐预测
// @Tags Eduadmin
// @Summary 用餐需求预测（weekday 均值）+ 食材采购与菜品调整建议
// @Security ApiKeyAuth
// @Produce application/json
// @Param date query string true "目标日期 yyyy-mm-dd"
// @Success 200 {object} response.Response{data=service.mealForecast} "预测完成"
// @Router /eduadmin/meal/forecast [get]
func (a *EduadminApi) ForecastEduMeals(c *gin.Context) {
	out, err := service.EduadminApp.ForecastEduMeals(c.Query("date"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "预测完成", c)
}
