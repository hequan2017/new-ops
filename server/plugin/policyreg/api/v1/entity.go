// Package v1 policyreg API：实体构建（需求 3.8-5）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"

	"github.com/hequan2017/new-ops/server/plugin/policyreg/service"
)

// ExtractPolicyEntities 实体抽取
// @Tags Policyreg
// @Summary 制度实体抽取（机构 词典+后缀规则/文件号/日期，按 docId 幂等覆盖）
// @Security ApiKeyAuth
// @Produce application/json
// @Param docId query int true "制度ID"
// @Success 200 {object} response.Response{data=[]model.PolicyEntity} "抽取完成"
// @Router /policyreg/entity/extract [post]
func (a *PolicyregApi) ExtractPolicyEntities(c *gin.Context) {
	docId, _ := strconv.ParseUint(c.Query("docId"), 10, 64)
	out, err := service.PolicyApp.ExtractPolicyEntities(uint(docId))
	if err != nil {
		errx.Respond(c, err, "操作失败")
		return
	}
	response.OkWithDetailed(out, "抽取完成", c)
}

// GetPolicyEntities 实体列表
// @Tags Policyreg
// @Summary 制度实体列表（类型过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param docId query int true "制度ID"
// @Param eType query string false "机构/文件号/日期/专有名词"
// @Success 200 {object} response.Response{data=[]model.PolicyEntity} "获取成功"
// @Router /policyreg/entity/list [get]
func (a *PolicyregApi) GetPolicyEntities(c *gin.Context) {
	docId, _ := strconv.ParseUint(c.Query("docId"), 10, 64)
	list, err := service.PolicyApp.GetPolicyEntities(uint(docId), c.Query("eType"))
	if err != nil {
		errx.Respond(c, err, "操作失败")
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
