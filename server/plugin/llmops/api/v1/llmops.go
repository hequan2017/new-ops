// Package v1 llmops API：大模型应用开发平台接口层（参数绑定与响应，禁止业务逻辑）
package v1

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/llmops/service"
)

var Api = new(LlmopsApi)

type LlmopsApi struct{}

// Ping 插件连通性探测
// @Tags Llmops
// @Summary llmops 连通性探测
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]any} "探测成功"
// @Router /llmops/ping [get]
func (a *LlmopsApi) Ping(c *gin.Context) {
	response.OkWithDetailed(gin.H{
		"plugin": "llmops",
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}, "探测成功", c)
}

// GetAssets 数据资产列表
// @Tags Llmops
// @Summary 数据资产列表（类型/状态/关键字过滤，白名单校验）
// @Security ApiKeyAuth
// @Produce application/json
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param dataType query string false "数据类型(doc/image/audio/video/webpage/qa/structured/semistructured)"
// @Param status query string false "状态(draft/parsed/cataloged/offline)"
// @Param keyword query string false "名称/标签关键字"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /llmops/asset/list [get]
func (a *LlmopsApi) GetAssets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	list, total, err := service.LlmopsService.GetAssetList(page, pageSize,
		c.Query("dataType"), c.Query("status"), c.Query("keyword"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetAsset 数据资产详情
// @Tags Llmops
// @Summary 数据资产详情
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "资产ID"
// @Success 200 {object} response.Response{data=model.LlmopsDataAsset} "获取成功"
// @Router /llmops/asset [get]
func (a *LlmopsApi) GetAsset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	asset, err := service.LlmopsService.GetAsset(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(asset, "获取成功", c)
}

// DeleteAsset 删除数据资产（软删除）
// @Tags Llmops
// @Summary 删除数据资产（软删除）
// @Security ApiKeyAuth
// @Param id query int true "资产ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/asset [delete]
func (a *LlmopsApi) DeleteAsset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteAsset(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
