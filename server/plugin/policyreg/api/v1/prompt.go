// Package v1 policyreg API：提示词库与问答（需求 3.8-8）
package v1

import (

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
	"github.com/hequan2017/new-ops/server/plugin/policyreg/service"
)

// SavePolicyPrompt 保存模板
// @Tags Policyreg
// @Summary 新建/更新制度问答模板（vars 自动解析，按名幂等）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.PolicyPromptTemplate true "模板"
// @Success 200 {object} response.Response{data=model.PolicyPromptTemplate} "保存成功"
// @Router /policyreg/prompt [post]
func (a *PolicyregApi) SavePolicyPrompt(c *gin.Context) {
	var in model.PolicyPromptTemplate
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.PolicyApp.SavePolicyPrompt(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetPolicyPrompts 模板列表
// @Tags Policyreg
// @Summary 提示词模板列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.PolicyPromptTemplate} "获取成功"
// @Router /policyreg/prompt/list [get]
func (a *PolicyregApi) GetPolicyPrompts(c *gin.Context) {
	list, err := service.PolicyApp.GetPolicyPrompts()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// qaReq 问答请求
type qaReq struct {
	RoleName  string `json:"roleName" binding:"required"`
	Question  string `json:"question" binding:"required"`
	ModelId   uint   `json:"modelId"`
}

// PolicyQa 制度问答
// @Tags Policyreg
// @Summary 制度问答（角色过滤检索 top3+模板注入+chat 可选，未配置模型规则口径兜底）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.qaReq true "问答请求"
// @Success 200 {object} response.Response{data=model.PolicyQaLog} "回答完成"
// @Router /policyreg/qa [post]
func (a *PolicyregApi) PolicyQa(c *gin.Context) {
	var in qaReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.PolicyApp.PolicyQa(in.RoleName, in.Question, in.ModelId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "回答完成", c)
}

// RenderPolicyPromptByName 模板渲染
// @Tags Policyreg
// @Summary 按模板名渲染（缺失占位变量报错）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.renderPrReq true "渲染请求"
// @Success 200 {object} response.Response{data=map[string]any} "渲染成功"
// @Router /policyreg/prompt/render [post]
func (a *PolicyregApi) RenderPolicyPromptByName(c *gin.Context) {
	var in struct {
		Name string            `json:"name" binding:"required"`
		Vars map[string]string `json:"vars"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.PolicyApp.RenderPolicyPromptByName(in.Name, in.Vars)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "渲染成功", c)
}
