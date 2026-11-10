// Package v1 eduadmin API：智能采购（需求 3.6-19）
package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SaveEduPurchase 保存采购记录
// @Tags Eduadmin
// @Summary 新增历史采购记录（按月）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduPurchase true "采购记录"
// @Success 200 {object} response.Response{data=model.EduPurchase} "保存成功"
// @Router /eduadmin/purchase [post]
func (a *EduadminApi) SaveEduPurchase(c *gin.Context) {
	var in model.EduPurchase
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduPurchase(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetEduPurchases 采购记录列表
// @Tags Eduadmin
// @Summary 采购记录列表（品目过滤，年月升序）
// @Security ApiKeyAuth
// @Produce application/json
// @Param item query string false "品目"
// @Success 200 {object} response.Response{data=[]model.EduPurchase} "获取成功"
// @Router /eduadmin/purchase/list [get]
func (a *EduadminApi) GetEduPurchases(c *gin.Context) {
	list, err := service.EduadminApp.GetEduPurchases(c.Query("item"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SuggestEduPurchase 采购建议
// @Tags Eduadmin
// @Summary 智能采购建议（移动平均预测量+峰值月提前 1 月下单+预算测算）
// @Security ApiKeyAuth
// @Produce application/json
// @Param item query string true "品目"
// @Success 200 {object} response.Response{data=service.purchaseSuggestion} "建议完成"
// @Router /eduadmin/purchase/suggest [get]
func (a *EduadminApi) SuggestEduPurchase(c *gin.Context) {
	out, err := service.EduadminApp.SuggestEduPurchase(c.Query("item"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "建议完成", c)
}
