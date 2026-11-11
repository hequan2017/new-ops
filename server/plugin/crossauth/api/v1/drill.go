// Package v1 crossauth API：应急演练与预案（需求 3.7-16）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// SaveCaPlaybook 保存预案
// @Tags Crossauth
// @Summary 新建/更新应急预案（更新即版本+1，按事件名幂等）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.CaPlaybook true "预案"
// @Success 200 {object} response.Response{data=model.CaPlaybook} "保存成功"
// @Router /crossauth/playbook [post]
func (a *CrossauthApi) SaveCaPlaybook(c *gin.Context) {
	var in model.CaPlaybook
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.SaveCaPlaybook(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetCaPlaybooks 预案列表
// @Tags Crossauth
// @Summary 应急预案列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.CaPlaybook} "获取成功"
// @Router /crossauth/playbook/list [get]
func (a *CrossauthApi) GetCaPlaybooks(c *gin.Context) {
	list, err := service.CaApp.GetCaPlaybooks()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveCaDrillPlan 演练计划
// @Tags Crossauth
// @summary 新建演练计划（同预案同年度唯一，年度 2 次口径字段）
// @Tags Crossauth
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.CaDrillPlan true "演练计划"
// @Success 200 {object} response.Response{data=model.CaDrillPlan} "保存成功"
// @Router /crossauth/drill [post]
func (a *CrossauthApi) SaveCaDrillPlan(c *gin.Context) {
	var in model.CaDrillPlan
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.SaveCaDrillPlan(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// FinishCaDrillPlan 复盘
// @Tags Crossauth
// @Summary 演练复盘归档（报告+评定 优/良/需改进）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.finishDrillReq true "复盘"
// @Success 200 {object} response.Response{data=model.CaDrillPlan} "归档成功"
// @Router /crossauth/drill/finish [post]
func (a *CrossauthApi) FinishCaDrillPlan(c *gin.Context) {
	var in finishDrillReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.FinishCaDrillPlan(in.ID, in.Report, in.Grade)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "归档成功", c)
}

// finishDrillReq 复盘请求
type finishDrillReq struct {
	ID     uint   `json:"id" binding:"required"`
	Report string `json:"report" binding:"required"`
	Grade  string `json:"grade" binding:"required"`
}

// GetCaDrillPlans 计划列表
// @Tags Crossauth
// @Summary 演练计划列表（年度过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param year query int false "年度"
// @Success 200 {object} response.Response{data=[]model.CaDrillPlan} "获取成功"
// @Router /crossauth/drill/list [get]
func (a *CrossauthApi) GetCaDrillPlans(c *gin.Context) {
	year, _ := strconv.Atoi(c.Query("year"))
	list, err := service.CaApp.GetCaDrillPlans(year)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
