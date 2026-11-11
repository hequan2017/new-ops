// Package v1 policyreg API：模型鉴权知识库（需求 3.8-7）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
	"github.com/hequan2017/new-ops/server/plugin/policyreg/service"
)

// SavePolicyPermGrant 保存 grant
// @Tags Policyreg
// @Summary 保存角色×分类权限 grant（可访问最高密级，按角色+分类幂等）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.PolicyPermGrant true "grant"
// @Success 200 {object} response.Response{data=model.PolicyPermGrant} "保存成功"
// @Router /policyreg/acl [post]
func (a *PolicyregApi) SavePolicyPermGrant(c *gin.Context) {
	var in model.PolicyPermGrant
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.PolicyApp.SavePolicyPermGrant(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetPolicyPermGrants grant 列表
// @Tags Policyreg
// @Summary 权限 grant 列表（角色过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param roleName query string false "角色"
// @Success 200 {object} response.Response{data=[]model.PolicyPermGrant} "获取成功"
// @Router /policyreg/acl/list [get]
func (a *PolicyregApi) GetPolicyPermGrants(c *gin.Context) {
	list, err := service.PolicyApp.GetPolicyPermGrants(c.Query("roleName"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SearchPolicyDocsByRole 按角色检索
// @Tags Policyreg
// @Summary 制度检索（角色过滤：无 grant/超密级/Allowed=false 不可见）
// @Security ApiKeyAuth
// @Produce application/json
// @Param roleName query string true "角色"
// @Param keyword query string false "名称关键字"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "检索成功"
// @Router /policyreg/acl/search [get]
func (a *PolicyregApi) SearchPolicyDocsByRole(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.PolicyApp.SearchPolicyDocsByRole(c.Query("roleName"), c.Query("keyword"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "检索成功", c)
}
