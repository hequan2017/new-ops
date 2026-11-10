// Package v1 eduadmin API：六大知识库（需求 3.6-4）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SaveEduPolicy 保存知识库条目
// @Tags Eduadmin
// @Summary 新建/更新知识库条目（六域白名单）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduPolicy true "条目"
// @Success 200 {object} response.Response{data=model.EduPolicy} "保存成功"
// @Router /eduadmin/policy [post]
func (a *EduadminApi) SaveEduPolicy(c *gin.Context) {
	var in model.EduPolicy
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduPolicy(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeleteEduPolicy 删除条目
// @Tags Eduadmin
// @Summary 删除知识库条目（软删除）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "条目ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /eduadmin/policy [delete]
func (a *EduadminApi) DeleteEduPolicy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.EduadminApp.DeleteEduPolicy(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetEduPolicies 条目列表
// @Tags Eduadmin
// @Summary 知识库条目列表（域过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param domain query string false "人事/财务/资产/后勤/决策/跨校"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/policy/list [get]
func (a *EduadminApi) GetEduPolicies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.EduadminApp.GetEduPolicies(c.Query("domain"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// SearchEduPolicies 知识库检索
// @Tags Eduadmin
// @Summary 知识库 bigram 检索（域过滤，命中降序 top-k）
// @Security ApiKeyAuth
// @Produce application/json
// @Param query query string true "检索词"
// @Param domain query string false "域"
// @Param topK query int false "返回条数"
// @Success 200 {object} response.Response{data=[]service.policyHit} "检索成功"
// @Router /eduadmin/policy/search [get]
func (a *EduadminApi) SearchEduPolicies(c *gin.Context) {
	topK, _ := strconv.Atoi(c.DefaultQuery("topK", "5"))
	hits, err := service.EduadminApp.SearchEduPolicies(c.Query("query"), c.Query("domain"), topK)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(hits, "检索成功", c)
}
