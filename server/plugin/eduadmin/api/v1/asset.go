// Package v1 eduadmin API：资产故障预测（需求 3.6-18）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SaveEduAsset 保存设备
// @Tags Eduadmin
// @Summary 新建/更新设备资产（指标+阈值）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduAsset true "设备"
// @Success 200 {object} response.Response{data=model.EduAsset} "保存成功"
// @Router /eduadmin/asset [post]
func (a *EduadminApi) SaveEduAsset(c *gin.Context) {
	var in model.EduAsset
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduAsset(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetEduAssets 设备列表
// @Tags Eduadmin
// @Summary 设备资产列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.EduAsset} "获取成功"
// @Router /eduadmin/asset/list [get]
func (a *EduadminApi) GetEduAssets(c *gin.Context) {
	list, err := service.EduadminApp.GetEduAssets()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// AddEduSensorReading 新增读数
// @Tags Eduadmin
// @Summary 新增传感器时序读数
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduSensorReading true "读数"
// @Success 200 {object} response.Response{data=model.EduSensorReading} "新增成功"
// @Router /eduadmin/asset/reading [post]
func (a *EduadminApi) AddEduSensorReading(c *gin.Context) {
	var in model.EduSensorReading
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.AddEduSensorReading(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "新增成功", c)
}

// PredictAssetFault 故障预测
// @Tags Eduadmin
// @Summary 资产故障预测（趋势斜率+阈值双因子判级+维护计划，按设备幂等）
// @Security ApiKeyAuth
// @Produce application/json
// @Param assetId query int true "设备ID"
// @Success 200 {object} response.Response{data=model.EduMaintPlan} "预测完成"
// @Router /eduadmin/asset/maintenance [get]
func (a *EduadminApi) PredictAssetFault(c *gin.Context) {
	assetId, _ := strconv.ParseUint(c.Query("assetId"), 10, 64)
	out, err := service.EduadminApp.PredictAssetFault(uint(assetId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "预测完成", c)
}

// GetEduMaintPlans 维护计划列表
// @Tags Eduadmin
// @Summary 维护计划列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.EduMaintPlan} "获取成功"
// @Router /eduadmin/asset/maintenance/list [get]
func (a *EduadminApi) GetEduMaintPlans(c *gin.Context) {
	list, err := service.EduadminApp.GetEduMaintPlans()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
