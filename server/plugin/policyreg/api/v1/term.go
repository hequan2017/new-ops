// Package v1 policyreg API：制度分词与属性（需求 3.8-4）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/policyreg/service"
)

// TagPolicyTerms 分词标注
// @Tags Policyreg
// @Summary 制度分词与属性标注（30 词词典最大正向匹配+分类标签+生效日期正则，按 docId 幂等）
// @Security ApiKeyAuth
// @Produce application/json
// @Param docId query int true "制度ID"
// @Success 200 {object} response.Response{data=model.PolicyTermTag} "标注完成"
// @Router /policyreg/term/tag [post]
func (a *PolicyregApi) TagPolicyTerms(c *gin.Context) {
	docId, _ := strconv.ParseUint(c.Query("docId"), 10, 64)
	out, err := service.PolicyApp.TagPolicyTerms(uint(docId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "标注完成", c)
}

// GetPolicyTermTags 标注列表
// @Tags Policyreg
// @Summary 分词标注列表（按制度）
// @Security ApiKeyAuth
// @Produce application/json
// @Param docId query int true "制度ID"
// @Success 200 {object} response.Response{data=[]model.PolicyTermTag} "获取成功"
// @Router /policyreg/term/tag/list [get]
func (a *PolicyregApi) GetPolicyTermTags(c *gin.Context) {
	docId, _ := strconv.ParseUint(c.Query("docId"), 10, 64)
	list, err := service.PolicyApp.GetPolicyTermTags(uint(docId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// GetPolicyTermStats 分类统计
// @Tags Policyreg
// @Summary 制度分类标签统计（六类固定序含零计数）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]service.LabelCount} "获取成功"
// @Router /policyreg/term/stats [get]
func (a *PolicyregApi) GetPolicyTermStats(c *gin.Context) {
	out, err := service.PolicyApp.GetPolicyTermStats()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "获取成功", c)
}
