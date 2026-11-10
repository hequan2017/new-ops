// Package v1 eduadmin API：人才预测（需求 3.6-14）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SaveEduTalentStat 保存年度统计
// @Tags Eduadmin
// @Summary 新建/更新年度人才统计（按年幂等）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduTalentStat true "年度统计"
// @Success 200 {object} response.Response{data=model.EduTalentStat} "保存成功"
// @Router /eduadmin/talent/stat [post]
func (a *EduadminApi) SaveEduTalentStat(c *gin.Context) {
	var in model.EduTalentStat
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduTalentStat(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetEduTalentStats 年度统计列表
// @Tags Eduadmin
// @Summary 年度人才统计列表（升序）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.EduTalentStat} "获取成功"
// @Router /eduadmin/talent/stat/list [get]
func (a *EduadminApi) GetEduTalentStats(c *gin.Context) {
	list, err := service.EduadminApp.GetEduTalentStats()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// ForecastTalent 人才预测
// @Tags Eduadmin
// @Summary 人才需求预测（线性回归正规方程）+ 流失率移动平均 + 培训计划建议
// @Security ApiKeyAuth
// @Produce application/json
// @Param years query int false "预测年数(默认3)"
// @Success 200 {object} response.Response{data=service.TalentForecast} "预测完成"
// @Router /eduadmin/talent/forecast [get]
func (a *EduadminApi) ForecastTalent(c *gin.Context) {
	years, _ := strconv.Atoi(c.DefaultQuery("years", "3"))
	out, err := service.EduadminApp.ForecastTalent(years)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "预测完成", c)
}
