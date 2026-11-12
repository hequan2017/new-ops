// Package v1 policyreg API：训练数据标注（需求 3.8-6）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"

	"github.com/hequan2017/new-ops/server/plugin/policyreg/service"
)

// BuildPolicyLabels 预标
// @Tags Policyreg
// @Summary 训练数据规则预标（实体回标分词+机构×文件号关系，先清后插幂等）
// @Security ApiKeyAuth
// @Produce application/json
// @Param docId query int true "制度ID"
// @Success 200 {object} response.Response{data=map[string]any} "预标完成"
// @Router /policyreg/label/pre [post]
func (a *PolicyregApi) BuildPolicyLabels(c *gin.Context) {
	docId, _ := strconv.ParseUint(c.Query("docId"), 10, 64)
	out, err := service.PolicyApp.BuildPolicyLabels(uint(docId))
	if err != nil {
		errx.Respond(c, err, "操作失败")
		return
	}
	response.OkWithDetailed(out, "预标完成", c)
}

// tokenLabelReq 人工改标请求
type tokenLabelReq struct {
	ID    uint   `json:"id" binding:"required"`
	Label string `json:"label" binding:"required"`
}

// UpdatePolicyTokenLabel 人工改标
// @Tags Policyreg
// @Summary 分词标注人工改标（O/机构/文件号/日期/专有名词，置人工来源）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.tokenLabelReq true "改标"
// @Success 200 {object} response.Response{data=map[string]any} "改标成功"
// @Router /policyreg/label/token [post]
func (a *PolicyregApi) UpdatePolicyTokenLabel(c *gin.Context) {
	var in tokenLabelReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.PolicyApp.UpdatePolicyTokenLabel(in.ID, in.Label)
	if err != nil {
		errx.Respond(c, err, "操作失败")
		return
	}
	response.OkWithDetailed(out, "改标成功", c)
}

// GetPolicyTokenLabels 分词标注列表
// @Tags Policyreg
// @Summary 分词标注列表（按制度，词序升序）
// @Security ApiKeyAuth
// @Produce application/json
// @Param docId query int true "制度ID"
// @Success 200 {object} response.Response{data=[]model.PolicyTokenLabel} "获取成功"
// @Router /policyreg/label/list [get]
func (a *PolicyregApi) GetPolicyTokenLabels(c *gin.Context) {
	docId, _ := strconv.ParseUint(c.Query("docId"), 10, 64)
	list, err := service.PolicyApp.GetPolicyTokenLabels(uint(docId))
	if err != nil {
		errx.Respond(c, err, "操作失败")
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// GetPolicyRelationLabels 关系标注列表
// @Tags Policyreg
// @summary 关系标注列表（按制度）
// @Tags Policyreg
// @Security ApiKeyAuth
// @Produce application/json
// @Param docId query int true "制度ID"
// @Success 200 {object} response.Response{data=[]model.PolicyRelationLabel} "获取成功"
// @Router /policyreg/label/relations [get]
func (a *PolicyregApi) GetPolicyRelationLabels(c *gin.Context) {
	docId, _ := strconv.ParseUint(c.Query("docId"), 10, 64)
	list, err := service.PolicyApp.GetPolicyRelationLabels(uint(docId))
	if err != nil {
		errx.Respond(c, err, "操作失败")
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
