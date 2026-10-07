// Package api 产品线接口
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
)

// parseUint 安全解析 uint 查询参数
func parseUint(s string) uint {
	v, _ := strconv.ParseUint(s, 10, 64)
	return uint(v)
}

type assetProductLine struct{}

// CreateAssetProductLine 创建产品线
// @Tags AssetProductLine
// @Summary 创建产品线
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.AssetProductLine true "产品线"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /asset/productLine/create [post]
func (a *assetProductLine) CreateAssetProductLine(c *gin.Context) {
	var p model.AssetProductLine
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetProductLineService.CreateAssetProductLine(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteAssetProductLine 删除产品线
// @Tags AssetProductLine
// @Summary 删除产品线
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.GetById true "产品线ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /asset/productLine/delete [delete]
func (a *assetProductLine) DeleteAssetProductLine(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetProductLineService.DeleteAssetProductLine(uint(req.ID)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// UpdateAssetProductLine 更新产品线
// @Tags AssetProductLine
// @Summary 更新产品线
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.AssetProductLine true "产品线（含ID）"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /asset/productLine/update [put]
func (a *assetProductLine) UpdateAssetProductLine(c *gin.Context) {
	var p model.AssetProductLine
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetProductLineService.UpdateAssetProductLine(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// GetAssetProductLineList 产品线全量列表
// @Tags AssetProductLine
// @Summary 产品线全量列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param keyword query string false "名称/负责人关键字"
// @Success 200 {object} response.Response{data=[]model.AssetProductLine} "获取成功"
// @Router /asset/productLine/list [get]
func (a *assetProductLine) GetAssetProductLineList(c *gin.Context) {
	list, err := assetProductLineService.GetAssetProductLineList(c.Query("keyword"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}
