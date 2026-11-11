// Package v1 eduadmin API：政策模拟（需求 3.6-26）
package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SimulateEduPolicy 单方案模拟
// @Tags Eduadmin
// @Summary 决策场景财务影响模拟（预算增减+费率调整，弹性系数模型，风险三档）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body service.SimInput true "场景参数"
// @Success 200 {object} response.Response{data=service.SimResult} "模拟完成"
// @Router /eduadmin/policy/simulate [post]
func (a *EduadminApi) SimulateEduPolicy(c *gin.Context) {
	var in service.SimInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SimulateEduPolicy(in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "模拟完成", c)
}

// compareSimReq 双方案请求
type compareSimReq struct {
	PlanA service.SimInput `json:"planA" binding:"required"`
	PlanB service.SimInput `json:"planB" binding:"required"`
}

// CompareEduPolicySims 双方案对比
// @Tags Eduadmin
// @Summary 两套政策方案模拟对比（净影响差额+风险等级摘要）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.compareSimReq true "双方案参数"
// @Success 200 {object} response.Response{data=map[string]any} "对比完成"
// @Router /eduadmin/policy/compare [post]
func (a *EduadminApi) CompareEduPolicySims(c *gin.Context) {
	var in compareSimReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	resA, resB, summary, err := service.EduadminApp.CompareEduPolicySims(in.PlanA, in.PlanB)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"planA": resA, "planB": resB, "summary": summary}, "对比完成", c)
}
