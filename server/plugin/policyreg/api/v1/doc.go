// Package v1 policyreg API：制度导入（需求 3.8-1/2）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
	"github.com/hequan2017/new-ops/server/plugin/policyreg/service"
)

// SavePolicyDoc 导入制度
// @Tags Policyreg
// @Summary 导入制度文档（TXT/MD 直读，Word/PDF/HTML 编目口径同 A03；涉密自动脱敏+SM4 口径标注；幂等）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.PolicyDoc true "制度"
// @Success 200 {object} response.Response{data=model.PolicyDoc} "导入成功"
// @Router /policyreg/doc [post]
func (a *PolicyregApi) SavePolicyDoc(c *gin.Context) {
	var in model.PolicyDoc
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.PolicyApp.SavePolicyDoc(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "导入成功", c)
}

// GetPolicyDocs 制度列表
// @Tags Policyreg
// @Summary 制度文档列表（格式/密级过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param format query string false "Word/PDF/HTML/TXT/MD"
// @Param securityLevel query string false "公开/内部/涉密"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /policyreg/doc/list [get]
func (a *PolicyregApi) GetPolicyDocs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.PolicyApp.GetPolicyDocs(c.Query("format"), c.Query("securityLevel"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetPolicyDoc 制度详情
// @Tags Policyreg
// @Summary 制度文档详情（按 ID）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "制度ID"
// @Success 200 {object} response.Response{data=model.PolicyDoc} "获取成功"
// @Router /policyreg/doc [get]
func (a *PolicyregApi) GetPolicyDoc(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	out, err := service.PolicyApp.GetPolicyDoc(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "获取成功", c)
}
