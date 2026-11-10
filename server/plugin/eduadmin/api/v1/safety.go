// Package v1 eduadmin API：校园安防-环境监测（需求 3.6-23）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SaveEduEnvReading 登记读数
// @Tags Eduadmin
// @Summary 新增环境监测读数（空气/水质，中/高自动告警+健康建议）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduEnvReading true "读数"
// @Success 200 {object} response.Response{data=model.EduEnvReading} "保存成功"
// @Router /eduadmin/env [post]
func (a *EduadminApi) SaveEduEnvReading(c *gin.Context) {
	var in model.EduEnvReading
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduEnvReading(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetEduEnvReadings 读数列表
// @Tags Eduadmin
// @Summary 环境读数列表（类型过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param type query string false "空气/水质"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/env/list [get]
func (a *EduadminApi) GetEduEnvReadings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.EduadminApp.GetEduEnvReadings(c.Query("type"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetEduEnvAlerts 告警列表
// @Tags Eduadmin
// @Summary 环境告警列表（含健康建议）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.EduEnvAlert} "获取成功"
// @Router /eduadmin/env/alerts [get]
func (a *EduadminApi) GetEduEnvAlerts(c *gin.Context) {
	list, err := service.EduadminApp.GetEduEnvAlerts()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
