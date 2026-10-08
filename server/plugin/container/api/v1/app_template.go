// Package api 白泽容器管理：应用模板一键部署接口（M8 C3）
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/container/model"
	"github.com/hequan2017/new-ops/server/plugin/container/service"
	"github.com/hequan2017/new-ops/server/utils"
)

// GetAppTemplates 模板列表
// @Tags DockerAppTemplate
// @Summary 应用模板列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.DockerAppTemplate} "获取成功"
// @Router /container/appTemplate/list [get]
func (a *containerApi) GetAppTemplates(c *gin.Context) {
	list, err := service.Service.Compose.GetAppTemplates()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveAppTemplate 保存模板
// @Tags DockerAppTemplate
// @Summary 新建/更新应用模板
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.DockerAppTemplate true "模板"
// @Success 200 {object} response.Response{msg=string} "保存成功"
// @Router /container/appTemplate [post]
func (a *containerApi) SaveAppTemplate(c *gin.Context) {
	var t model.DockerAppTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.Service.Compose.SaveAppTemplate(&t); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// DeleteAppTemplate 删除模板
// @Tags DockerAppTemplate
// @Summary 删除应用模板
// @Security ApiKeyAuth
// @Param id query int true "模板ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /container/appTemplate [delete]
func (a *containerApi) DeleteAppTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := service.Service.Compose.DeleteAppTemplate(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeployFromTemplate 一键部署
// @Tags DockerAppTemplate
// @Summary 模板渲染 compose 后一键部署
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body object true "templateId/endpointId/projectName/params"
// @Success 200 {object} response.Response{msg=string} "部署成功"
// @Router /container/appTemplate/deploy [post]
func (a *containerApi) DeployFromTemplate(c *gin.Context) {
	var req struct {
		TemplateID  uint              `json:"templateId" binding:"required"`
		EndpointID  uint              `json:"endpointId" binding:"required"`
		ProjectName string            `json:"projectName" binding:"required"`
		Params      map[string]string `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	proj, err := service.Service.Compose.DeployFromTemplate(
		req.TemplateID, req.EndpointID, req.ProjectName, req.Params, utils.GetUserName(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(proj, "部署完成（up 失败时项目已留存，可查看状态后重试）", c)
}

// GetAppInstances 模板部署实例列表
// @Tags DockerAppTemplate
// @Summary 从模板部署的应用实例列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]map[string]any} "获取成功"
// @Router /container/appTemplate/instances [get]
func (a *containerApi) GetAppInstances(c *gin.Context) {
	list, err := service.Service.Compose.GetAppInstances()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
