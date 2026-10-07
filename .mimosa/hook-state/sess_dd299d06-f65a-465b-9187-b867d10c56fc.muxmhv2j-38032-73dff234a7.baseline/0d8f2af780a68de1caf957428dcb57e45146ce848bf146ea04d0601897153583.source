// Package api 主机资产接口（Swagger 注释按 DEV_PLAN 3.6：/asset/host 前缀）
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
	assetReq "github.com/hequan2017/new-ops/server/plugin/asset/model/request"
	"github.com/hequan2017/new-ops/server/plugin/asset/service"
	"github.com/hequan2017/new-ops/server/utils"
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
	if err := assetHostService.CreateAssetHost(&h, utils.GetUserName(c)); err != nil {
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
	if err := assetHostService.DeleteAssetHost(uint(req.ID), utils.GetUserName(c)); err != nil {
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
	if err := assetHostService.DeleteAssetHostByIds(ids, utils.GetUserName(c)); err != nil {
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
	if err := assetHostService.UpdateAssetHost(&h, utils.GetUserName(c)); err != nil {
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
	scope := &service.HostUserScope{
		IsSuperAdmin: utils.GetUserAuthorityId(c) == 888,
		UserID:       utils.GetUserID(c),
	}
	list, total, err := assetHostService.GetAssetHostList(req.PageInfo, req.Status, req.RoomID, req.ProductLineID, scope)
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

// GetAssetHostHistory 分页查询主机资产变更历史
// @Tags AssetHost
// @Summary 分页查询主机资产变更历史
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "主机ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /asset/host/history [get]
func (a *assetHost) GetAssetHostHistory(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var info request.PageInfo
	if err := c.ShouldBindQuery(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := assetHostService.GetAssetHostHistoryList(uint(req.ID), info)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     info.Page,
		PageSize: info.PageSize,
	}, "获取成功", c)
}

// ExportAssetHost 导出主机资产 Excel
// @Tags AssetHost
// @Summary 导出主机资产 Excel
// @Security ApiKeyAuth
// @Produce application/octet-stream
// @Success 200 {file} file "Excel 文件流"
// @Router /asset/host/export [get]
func (a *assetHost) ExportAssetHost(c *gin.Context) {
	f, err := assetHostService.ExportAssetHostsForUser(&service.HostUserScope{
		IsSuperAdmin: utils.GetUserAuthorityId(c) == 888,
		UserID:       utils.GetUserID(c),
	})
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=asset_hosts.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// ImportAssetHost 导入主机资产 Excel（按内网IP upsert）
// @Tags AssetHost
// @Summary 导入主机资产 Excel
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce application/json
// @Param file formData file true "Excel 文件"
// @Success 200 {object} response.Response{data=service.ImportResult} "导入完成"
// @Router /asset/host/import [post]
func (a *assetHost) ImportAssetHost(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("请上传文件", c)
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	defer f.Close()
	res, err := assetHostService.ImportAssetHosts(f, utils.GetUserName(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(res, "导入完成", c)
}

// CollectAssetHost 用绑定的 SSH 凭据现场采集主机信息并回填
// @Tags AssetHost
// @Summary SSH 现场采集主机信息
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "ID(主机ID)/credentialId(SSH凭据ID)"
// @Success 200 {object} response.Response{data=service.CollectedInfo,msg=string} "采集成功"
// @Router /asset/host/collect [post]
func (a *assetHost) CollectAssetHost(c *gin.Context) {
	var req struct {
		ID          uint `json:"ID" binding:"required"`
		CredentialID uint `json:"credentialId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	info, err := assetHostService.CollectHostFromCredential(req.ID, req.CredentialID, utils.GetUserName(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(info, "采集成功", c)
}

// SyncAliyunECS 从阿里云 ECS 同步实例到资产台账
// @Tags AssetHost
// @Summary 阿里云 ECS 实例同步
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "credentialId(cloud_ak凭据)/region(如 cn-beijing)"
// @Success 200 {object} response.Response{data=service.ECSSyncResult,msg=string} "同步完成"
// @Router /asset/host/sync/aliyun-ecs [post]
func (a *assetHost) SyncAliyunECS(c *gin.Context) {
	var req struct {
		CredentialID uint   `json:"credentialId" binding:"required"`
		Region       string `json:"region" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	res, err := assetHostService.SyncFromAliyunECS(req.CredentialID, req.Region, utils.GetUserName(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(res, "同步完成", c)
}

// DiscoverHosts CIDR 网段 SSH 服务探测
// @Tags AssetHost
// @Summary 网段发现（探测 CIDR 内开放 SSH 的主机）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "cidr(如 192.168.1.0/24)/port(默认22)/concurrency(默认50)/timeoutMs(默认1500)"
// @Success 200 {object} response.Response{data=[]service.DiscoveredHost} "探测完成"
// @Router /asset/host/discover [post]
func (a *assetHost) DiscoverHosts(c *gin.Context) {
	var req struct {
		CIDR        string `json:"cidr" binding:"required"`
		Port        int    `json:"port"`
		Concurrency int    `json:"concurrency"`
		TimeoutMs   int    `json:"timeoutMs"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	// 探测整体限时，防超大段长阻塞
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	list, err := assetHostService.DiscoverHosts(ctx, req.CIDR, req.Port, req.Concurrency, req.TimeoutMs)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "探测完成", c)
}

// ImportDiscoveredHosts 导入发现的资产
// @Tags AssetHost
// @Summary 导入网段发现的主机（已存在IP跳过）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "hosts:[{ip,banner}]"
// @Success 200 {object} response.Response{data=object} "导入完成（created/skipped）"
// @Router /asset/host/discover/import [post]
func (a *assetHost) ImportDiscoveredHosts(c *gin.Context) {
	var req struct {
		Hosts []service.DiscoveredHost `json:"hosts" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if len(req.Hosts) > 4096 {
		response.FailWithMessage("单次导入超过 4096 台上限", c)
		return
	}
	created, skipped, err := assetHostService.ImportDiscoveredHosts(req.Hosts, utils.GetUserName(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"created": created, "skipped": skipped}, "导入完成", c)
}
