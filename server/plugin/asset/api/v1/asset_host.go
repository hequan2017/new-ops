// Package api 主机资产接口（Swagger 注释按 DEV_PLAN 3.6：/asset/host 前缀）
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
	assetReq "github.com/hequan2017/new-ops/server/plugin/asset/model/request"
)

type assetHost struct{}

// CreateAssetHost 创建主机资产
// @Tags AssetHost
// @Summary 创建主机资产
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.AssetHost true "主机资产"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /asset/host/create [post]
func (a *assetHost) CreateAssetHost(c *gin.Context) {
	var h model.AssetHost
	if err := c.ShouldBindJSON(&h); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetHostService.CreateAssetHost(&h); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteAssetHost 删除主机资产
// @Tags AssetHost
// @Summary 删除主机资产
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.GetById true "主机ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /asset/host/delete [delete]
func (a *assetHost) DeleteAssetHost(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetHostService.DeleteAssetHost(uint(req.ID)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteAssetHostByIds 批量删除主机资产
// @Tags AssetHost
// @Summary 批量删除主机资产
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量ID"
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /asset/host/deleteByIds [delete]
func (a *assetHost) DeleteAssetHostByIds(c *gin.Context) {
	var req request.IdsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	ids := make([]uint, 0, len(req.Ids))
	for _, id := range req.Ids {
		ids = append(ids, uint(id))
	}
	if err := assetHostService.DeleteAssetHostByIds(ids); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateAssetHost 更新主机资产
// @Tags AssetHost
// @Summary 更新主机资产
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.AssetHost true "主机资产（含ID）"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /asset/host/update [put]
func (a *assetHost) UpdateAssetHost(c *gin.Context) {
	var h model.AssetHost
	if err := c.ShouldBindJSON(&h); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetHostService.UpdateAssetHost(&h); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindAssetHost 根据ID获取主机资产
// @Tags AssetHost
// @Summary 根据ID获取主机资产
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.GetById true "主机ID"
// @Success 200 {object} response.Response{data=model.AssetHost} "获取成功"
// @Router /asset/host/find [get]
func (a *assetHost) FindAssetHost(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	h, err := assetHostService.GetAssetHost(uint(req.ID))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(h, c)
}

// GetAssetHostList 分页获取主机资产列表
// @Tags AssetHost
// @Summary 分页获取主机资产列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body assetReq.AssetHostSearch true "页码/页大小/keyword/status/roomId/productLineId"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /asset/host/list [post]
func (a *assetHost) GetAssetHostList(c *gin.Context) {
	var req assetReq.AssetHostSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := assetHostService.GetAssetHostList(req.PageInfo, req.Status, req.RoomID, req.ProductLineID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "获取成功", c)
}
