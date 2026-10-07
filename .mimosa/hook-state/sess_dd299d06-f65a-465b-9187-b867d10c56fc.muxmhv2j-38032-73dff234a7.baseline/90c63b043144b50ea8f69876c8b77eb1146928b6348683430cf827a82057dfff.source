// Package api 机房/机柜接口
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
)

type assetRoom struct{}

// CreateAssetRoom 创建机房
// @Tags AssetRoom
// @Summary 创建机房
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.AssetRoom true "机房"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /asset/room/create [post]
func (a *assetRoom) CreateAssetRoom(c *gin.Context) {
	var r model.AssetRoom
	if err := c.ShouldBindJSON(&r); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetRoomService.CreateAssetRoom(&r); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteAssetRoom 删除机房
// @Tags AssetRoom
// @Summary 删除机房
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.GetById true "机房ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /asset/room/delete [delete]
func (a *assetRoom) DeleteAssetRoom(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetRoomService.DeleteAssetRoom(uint(req.ID)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// UpdateAssetRoom 更新机房
// @Tags AssetRoom
// @Summary 更新机房
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.AssetRoom true "机房（含ID）"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /asset/room/update [put]
func (a *assetRoom) UpdateAssetRoom(c *gin.Context) {
	var r model.AssetRoom
	if err := c.ShouldBindJSON(&r); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetRoomService.UpdateAssetRoom(&r); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// GetAssetRoomList 机房全量列表
// @Tags AssetRoom
// @Summary 机房全量列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param keyword query string false "名称/区域关键字"
// @Success 200 {object} response.Response{data=[]model.AssetRoom} "获取成功"
// @Router /asset/room/list [get]
func (a *assetRoom) GetAssetRoomList(c *gin.Context) {
	list, err := assetRoomService.GetAssetRoomList(c.Query("keyword"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

type assetRack struct{}

// CreateAssetRack 创建机柜
// @Tags AssetRack
// @Summary 创建机柜
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.AssetRack true "机柜"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /asset/rack/create [post]
func (a *assetRack) CreateAssetRack(c *gin.Context) {
	var r model.AssetRack
	if err := c.ShouldBindJSON(&r); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetRackService.CreateAssetRack(&r); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteAssetRack 删除机柜
// @Tags AssetRack
// @Summary 删除机柜
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.GetById true "机柜ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /asset/rack/delete [delete]
func (a *assetRack) DeleteAssetRack(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetRackService.DeleteAssetRack(uint(req.ID)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// UpdateAssetRack 更新机柜
// @Tags AssetRack
// @Summary 更新机柜
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.AssetRack true "机柜（含ID）"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /asset/rack/update [put]
func (a *assetRack) UpdateAssetRack(c *gin.Context) {
	var r model.AssetRack
	if err := c.ShouldBindJSON(&r); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetRackService.UpdateAssetRack(&r); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// GetAssetRackList 机柜列表（可按机房过滤）
// @Tags AssetRack
// @Summary 机柜列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param roomId query uint false "机房ID"
// @Success 200 {object} response.Response{data=[]model.AssetRack} "获取成功"
// @Router /asset/rack/list [get]
func (a *assetRack) GetAssetRackList(c *gin.Context) {
	var roomId *uint
	if v := c.Query("roomId"); v != "" {
		id := parseUint(v)
		roomId = &id
	}
	list, err := assetRackService.GetAssetRackList(roomId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}
