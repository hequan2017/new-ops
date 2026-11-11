// Package v1 crossauth API：数据分级引擎（需求 3.7-9）
package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// classifyScanReq 扫描请求
type classifyScanReq struct {
	Items []service.ClassifyScanItem `json:"items" binding:"required"`
}

// ScanCaClassify 批量扫描
// @Tags Crossauth
// @Summary 数据分级批量扫描（分类 教学/科研/管理 ×分级 公开/内部/敏感，关键词规则）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.classifyScanReq true "扫描请求"
// @Success 200 {object} response.Response{data=[]model.CaDataClassify} "扫描完成"
// @Router /crossauth/classify/scan [post]
func (a *CrossauthApi) ScanCaClassify(c *gin.Context) {
	var in classifyScanReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.ScanCaClassify(in.Items)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "扫描完成", c)
}

// GetCaClassifyResults 结果列表
// @Tags Crossauth
// @Summary 分级结果列表（分级/分类过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param level query string false "公开/内部/敏感"
// @Param category query string false "教学/科研/管理"
// @Success 200 {object} response.Response{data=[]model.CaDataClassify} "获取成功"
// @Router /crossauth/classify/list [get]
func (a *CrossauthApi) GetCaClassifyResults(c *gin.Context) {
	list, err := service.CaApp.GetCaClassifyResults(c.Query("level"), c.Query("category"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
