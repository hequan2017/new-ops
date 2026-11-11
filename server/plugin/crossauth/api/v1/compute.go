// Package v1 crossauth API：算力使用防护（需求 3.7-12）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// usageReq 用量请求
type usageReq struct {
	UserName  string  `json:"userName" binding:"required"`
	GpuHours  float64 `json:"gpuHours"`
	CoreHours float64 `json:"coreHours"`
}

// RecordCaComputeUsage 记录用量
// @Tags Crossauth
// @Summary 记录算力用量（用户×小时窗口累计；>1000 核时限流、>800 降优先级，越阈告警）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.usageReq true "用量"
// @Success 200 {object} response.Response{data=map[string]any} "记录完成"
// @Router /crossauth/compute/usage [post]
func (a *CrossauthApi) RecordCaComputeUsage(c *gin.Context) {
	var in usageReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.RecordCaComputeUsage(in.UserName, in.GpuHours, in.CoreHours)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "记录完成", c)
}

// GetCaComputeUsages 用量列表
// @Tags Crossauth
// @Summary 算力用量列表（用户过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param userName query string false "用户"
// @Success 200 {object} response.Response{data=[]model.CaComputeUsage} "获取成功"
// @Router /crossauth/compute/usage/list [get]
func (a *CrossauthApi) GetCaComputeUsages(c *gin.Context) {
	list, err := service.CaApp.GetCaComputeUsages(c.Query("userName"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// GetCaComputeAlerts 告警列表
// @Tags Crossauth
// @Summary 算力告警列表（处置状态过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param handled query string false "yes/no"
// @Success 200 {object} response.Response{data=[]model.CaComputeAlert} "获取成功"
// @Router /crossauth/compute/alerts [get]
func (a *CrossauthApi) GetCaComputeAlerts(c *gin.Context) {
	list, err := service.CaApp.GetCaComputeAlerts(c.Query("handled"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// HandleCaComputeAlert 处置
// @Tags Crossauth
// @Summary 算力告警处置标记
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "告警ID"
// @Success 200 {object} response.Response "处置成功"
// @Router /crossauth/compute/handle [post]
func (a *CrossauthApi) HandleCaComputeAlert(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if err := service.CaApp.HandleCaComputeAlert(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("处置成功", c)
}
