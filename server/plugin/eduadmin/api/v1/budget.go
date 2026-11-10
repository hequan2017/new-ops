// Package v1 eduadmin API：智能预算（需求 3.6-15）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SaveEduBudget 保存预算
// @Tags Eduadmin
// @Summary 新建/更新部门年度预算（同年同部门幂等）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduBudget true "预算"
// @Success 200 {object} response.Response{data=model.EduBudget} "保存成功"
// @Router /eduadmin/budget [post]
func (a *EduadminApi) SaveEduBudget(c *gin.Context) {
	var in model.EduBudget
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduBudget(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetEduBudgets 预算列表
// @Tags Eduadmin
// @Summary 预算列表（年度过滤，年/部门升序）
// @Security ApiKeyAuth
// @Produce application/json
// @Param year query int false "年度"
// @Success 200 {object} response.Response{data=[]model.EduBudget} "获取成功"
// @Router /eduadmin/budget/list [get]
func (a *EduadminApi) GetEduBudgets(c *gin.Context) {
	year, _ := strconv.Atoi(c.Query("year"))
	list, err := service.EduadminApp.GetEduBudgets(year)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// AnalyzeEduBudget 同比分析
// @Tags Eduadmin
// @Summary 部门预算同比/环比趋势（逐年增长率）
// @Security ApiKeyAuth
// @Produce application/json
// @Param dept query string true "部门"
// @Success 200 {object} response.Response{data=[]service.budgetDelta} "分析完成"
// @Router /eduadmin/budget/analysis [get]
func (a *EduadminApi) AnalyzeEduBudget(c *gin.Context) {
	out, err := service.EduadminApp.AnalyzeEduBudget(c.Query("dept"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "分析完成", c)
}

// SuggestEduBudget 分配建议
// @Tags Eduadmin
// @Summary 预算分配比例建议（上年占比缩放+单部门下限 10% 重归一）
// @Security ApiKeyAuth
// @Produce application/json
// @Param total query float true "新年度总预算(万元)"
// @Success 200 {object} response.Response{data=[]service.deptAmount} "建议完成"
// @Router /eduadmin/budget/suggest [get]
func (a *EduadminApi) SuggestEduBudget(c *gin.Context) {
	total, _ := strconv.ParseFloat(c.Query("total"), 64)
	out, err := service.EduadminApp.SuggestEduBudget(total)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "建议完成", c)
}

// compareReq 方案对比请求
type compareReq struct {
	PlanA []service.DeptAmount `json:"planA" binding:"required"`
	PlanB []service.DeptAmount `json:"planB" binding:"required"`
}

// CompareEduBudgetPlans 方案对比
// @Tags Eduadmin
// @Summary 两套预算方案对比（逐部门差额+总额+倾斜摘要）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.compareReq true "对比请求"
// @Success 200 {object} response.Response{data=service.planDiff} "对比完成"
// @Router /eduadmin/budget/compare [post]
func (a *EduadminApi) CompareEduBudgetPlans(c *gin.Context) {
	var in compareReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.CompareEduBudgetPlans(toDeptAmounts(in.PlanA), toDeptAmounts(in.PlanB))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "对比完成", c)
}

func toDeptAmounts(list []service.DeptAmount) []service.DeptAmount {
	out := make([]service.DeptAmount, 0, len(list))
	for _, d := range list {
		out = append(out, service.DeptAmount{Dept: d.Dept, Amount: d.Amount})
	}
	return out
}
