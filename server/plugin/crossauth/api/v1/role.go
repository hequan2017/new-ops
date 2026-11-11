// Package v1 crossauth API：角色模板（需求 3.7-5）
package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// SaveCaRoleTemplate 保存模板
// @Tags Crossauth
// @Summary 新建/更新角色模板（权限 JSON 校验，按角色名幂等）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.CaRoleTemplate true "模板"
// @Success 200 {object} response.Response{data=model.CaRoleTemplate} "保存成功"
// @Router /crossauth/roleTemplate [post]
func (a *CrossauthApi) SaveCaRoleTemplate(c *gin.Context) {
	var in model.CaRoleTemplate
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.SaveCaRoleTemplate(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetCaRoleTemplates 模板列表
// @Tags Crossauth
// @Summary 角色模板列表（分类过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param category query string false "教学/科研/实训/管理/企业/审计/运维"
// @Success 200 {object} response.Response{data=[]model.CaRoleTemplate} "获取成功"
// @Router /crossauth/roleTemplate/list [get]
func (a *CrossauthApi) GetCaRoleTemplates(c *gin.Context) {
	list, err := service.CaApp.GetCaRoleTemplates(c.Query("category"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
