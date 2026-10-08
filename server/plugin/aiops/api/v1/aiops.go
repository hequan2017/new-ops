// Package v1 aiops API：LLM 供应商管理 / 网关对话 / Pod 诊断
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/aiops/model"
	"github.com/hequan2017/new-ops/server/plugin/aiops/service"
)

var Api = new(AiopsApi)

type AiopsApi struct{}

// GetProviders LLM 供应商列表
// @Tags Aiops
// @Summary 供应商列表（密钥不回显）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.AiopsLlmProvider} "获取成功"
// @Router /aiops/provider/list [get]
func (a *AiopsApi) GetProviders(c *gin.Context) {
	list, err := service.AiopsService.GetProviders()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveProvider 保存供应商
// @Tags Aiops
// @Summary 新建/更新供应商（密钥留空不改）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.AiopsLlmProvider true "供应商"
// @Success 200 {object} response.Response{msg=string} "保存成功"
// @Router /aiops/provider [post]
func (a *AiopsApi) SaveProvider(c *gin.Context) {
	var p model.AiopsLlmProvider
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.AiopsService.SaveProvider(&p); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// DeleteProvider 删除供应商
// @Tags Aiops
// @Summary 删除供应商
// @Security ApiKeyAuth
// @Param id query int true "供应商ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /aiops/provider [delete]
func (a *AiopsApi) DeleteProvider(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := service.AiopsService.DeleteProvider(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// Chat LLM 对话
// @Tags Aiops
// @Summary 经网关调用 LLM（OpenAI 兼容）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body object true "providerId(可选)/prompt/system(可选)"
// @Success 200 {object} response.Response{data=object} "调用成功"
// @Router /aiops/llm/chat [post]
func (a *AiopsApi) Chat(c *gin.Context) {
	var req struct {
		ProviderID uint   `json:"providerId"`
		Prompt     string `json:"prompt" binding:"required"`
		System     string `json:"system"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	content, provider, modelName, err := service.AiopsService.Chat(
		c.Request.Context(), req.ProviderID, req.System, req.Prompt)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{
		"content": content, "provider": provider, "model": modelName,
	}, "调用成功", c)
}

// Diagnose Pod 诊断
// @Tags Aiops
// @Summary Pod 异常诊断（状态+事件+日志 → LLM）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body object true "clusterId/cluster/namespace/pod"
// @Success 200 {object} response.Response{data=model.AiopsDiagnosis} "诊断完成"
// @Router /aiops/diagnose [post]
func (a *AiopsApi) Diagnose(c *gin.Context) {
	var req struct {
		ClusterID uint   `json:"clusterId" binding:"required"`
		Cluster   string `json:"cluster"`
		Namespace string `json:"namespace" binding:"required"`
		Pod       string `json:"pod" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	rec, err := service.AiopsService.DiagnosePod(
		c.Request.Context(), req.ClusterID, req.Cluster, req.Namespace, req.Pod)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(rec, "诊断完成", c)
}

// GetDiagnosisList 诊断历史
// @Tags Aiops
// @Summary 诊断历史列表（快照不回传）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body object true "page/pageSize"
// @Success 200 {object} response.Response{data=object} "获取成功"
// @Router /aiops/diagnosis/list [post]
func (a *AiopsApi) GetDiagnosisList(c *gin.Context) {
	var req struct {
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}
	list, total, err := service.AiopsService.GetDiagnosisList(req.Page, req.PageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetDiagnosis 诊断详情
// @Tags Aiops
// @Summary 诊断详情（含快照与结论）
// @Security ApiKeyAuth
// @Param id query int true "诊断ID"
// @Success 200 {object} response.Response{data=model.AiopsDiagnosis} "获取成功"
// @Router /aiops/diagnosis [get]
func (a *AiopsApi) GetDiagnosis(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	rec, err := service.AiopsService.GetDiagnosis(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(rec, "获取成功", c)
}
