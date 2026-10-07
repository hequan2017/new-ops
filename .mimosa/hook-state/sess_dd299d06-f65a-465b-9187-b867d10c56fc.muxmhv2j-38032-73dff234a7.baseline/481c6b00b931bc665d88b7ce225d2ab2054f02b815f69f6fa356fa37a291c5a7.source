// Package api 资产组接口（数据权限单元）
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
)

type assetGroup struct{}

// CreateAssetGroup 创建资产组
// @Tags AssetGroup
// @Summary 创建资产组
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "name/notes/hostIds/userIds"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /asset/group/create [post]
func (a *assetGroup) CreateAssetGroup(c *gin.Context) {
	var req struct {
		Name    string `json:"name" binding:"required"`
		Notes   string `json:"notes"`
		HostIDs []uint `json:"hostIds"`
		UserIDs []uint `json:"userIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	g := &model.AssetGroup{Name: req.Name, Notes: req.Notes}
	if err := assetGroupService.CreateAssetGroup(g, req.HostIDs, req.UserIDs); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteAssetGroup 删除资产组
// @Tags AssetGroup
// @Summary 删除资产组
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.GetById true "资产组ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /asset/group/delete [delete]
func (a *assetGroup) DeleteAssetGroup(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := assetGroupService.DeleteAssetGroup(uint(req.ID)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// UpdateAssetGroup 更新资产组（重写主机与用户关联）
// @Tags AssetGroup
// @Summary 更新资产组
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "ID/name/notes/hostIds/userIds"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /asset/group/update [put]
func (a *assetGroup) UpdateAssetGroup(c *gin.Context) {
	var req struct {
		ID      uint   `json:"ID" binding:"required"`
		Name    string `json:"name" binding:"required"`
		Notes   string `json:"notes"`
		HostIDs []uint `json:"hostIds"`
		UserIDs []uint `json:"userIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	g := &model.AssetGroup{Name: req.Name, Notes: req.Notes}
	g.ID = req.ID
	if err := assetGroupService.UpdateAssetGroup(g, req.HostIDs, req.UserIDs); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// GetAssetGroupList 资产组全量列表（含主机/用户关联）
// @Tags AssetGroup
// @Summary 资产组全量列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]service.AssetGroupWithRelations} "获取成功"
// @Router /asset/group/list [get]
func (a *assetGroup) GetAssetGroupList(c *gin.Context) {
	list, err := assetGroupService.GetAssetGroupList()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}
