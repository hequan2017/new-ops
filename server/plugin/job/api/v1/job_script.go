// Package api 脚本库与变量组接口
package api

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/job/model"
	"github.com/hequan2017/new-ops/server/plugin/job/service"
	"github.com/hequan2017/new-ops/server/utils"
)

type jobScript struct{}

var scriptSvc = new(service.ScriptService)
var varGroupSvc = new(service.VariableGroupService)

// ---------- 脚本库 ----------

// CreateScript 创建脚本
// @Tags JobScript
// @Summary 创建脚本（初始版本 1）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.JobScript true "名称/语言/内容"
// @Success 200 {object} response.Response "创建成功"
// @Router /job/script [post]
func (j *jobScript) CreateScript(c *gin.Context) {
	var sc model.JobScript
	if err := c.ShouldBindJSON(&sc); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := scriptSvc.CreateScript(&sc, utils.GetUserName(c)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateScript 更新脚本（内容变更版本递增并归档）
// @Tags JobScript
// @Summary 更新脚本
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.JobScript true "含ID"
// @Success 200 {object} response.Response "更新成功"
// @Router /job/script [put]
func (j *jobScript) UpdateScript(c *gin.Context) {
	var sc model.JobScript
	if err := c.ShouldBindJSON(&sc); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := scriptSvc.UpdateScript(&sc, utils.GetUserName(c)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteScript 删除脚本
// @Tags JobScript
// @Summary 删除脚本（联动清版本）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "脚本ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /job/script [delete]
func (j *jobScript) DeleteScript(c *gin.Context) {
	id, err := parseUintQuery(c.Query("id"))
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := scriptSvc.DeleteScript(id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetScriptList 脚本列表
// @Tags JobScript
// @Summary 脚本列表（keyword/language 过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param keyword query string false "名称/备注模糊"
// @Param language query string false "语言过滤"
// @Success 200 {object} response.Response{data=[]model.JobScript} "获取成功"
// @Router /job/script/list [get]
func (j *jobScript) GetScriptList(c *gin.Context) {
	list, err := scriptSvc.GetScriptList(c.Query("keyword"), c.Query("language"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// GetScriptVersions 脚本历史版本
// @Tags JobScript
// @Summary 脚本历史版本（新→旧，50 条）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "脚本ID"
// @Success 200 {object} response.Response{data=[]model.JobScriptVersion} "获取成功"
// @Router /job/script/versions [get]
func (j *jobScript) GetScriptVersions(c *gin.Context) {
	id, err := parseUintQuery(c.Query("id"))
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	list, err := scriptSvc.GetScriptVersions(id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// ---------- 变量组 ----------

// CreateVariableGroup 创建变量组
// @Tags JobVariableGroup
// @Summary 创建变量组
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.JobVariableGroup true "名称/变量JSON/资产组"
// @Success 200 {object} response.Response "创建成功"
// @Router /job/vargroup [post]
func (j *jobScript) CreateVariableGroup(c *gin.Context) {
	var vg model.JobVariableGroup
	if err := c.ShouldBindJSON(&vg); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := varGroupSvc.CreateVariableGroup(&vg); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateVariableGroup 更新变量组
// @Tags JobVariableGroup
// @Summary 更新变量组
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.JobVariableGroup true "含ID"
// @Success 200 {object} response.Response "更新成功"
// @Router /job/vargroup [put]
func (j *jobScript) UpdateVariableGroup(c *gin.Context) {
	var vg model.JobVariableGroup
	if err := c.ShouldBindJSON(&vg); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := varGroupSvc.UpdateVariableGroup(&vg); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteVariableGroup 删除变量组
// @Tags JobVariableGroup
// @Summary 删除变量组
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "变量组ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /job/vargroup [delete]
func (j *jobScript) DeleteVariableGroup(c *gin.Context) {
	id, err := parseUintQuery(c.Query("id"))
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := varGroupSvc.DeleteVariableGroup(id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetVariableGroupList 变量组列表
// @Tags JobVariableGroup
// @Summary 变量组列表（keyword 过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param keyword query string false "名称/备注模糊"
// @Success 200 {object} response.Response{data=[]model.JobVariableGroup} "获取成功"
// @Router /job/vargroup/list [get]
func (j *jobScript) GetVariableGroupList(c *gin.Context) {
	list, err := varGroupSvc.GetVariableGroupList(c.Query("keyword"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
