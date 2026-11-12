// Package v1 policyreg API：制度清洗（需求 3.8-3）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"

	"github.com/hequan2017/new-ops/server/plugin/policyreg/service"
)

// CleanPolicyDoc 制度清洗
// @Tags Policyreg
// @Summary 制度清洗（编码乱码率/全半角格式冲突/重复条号/条款冲突关键词对，dry 报告）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "制度ID"
// @Success 200 {object} response.Response{data=map[string]any} "清洗完成"
// @Router /policyreg/clean [post]
func (a *PolicyregApi) CleanPolicyDoc(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	out, err := service.PolicyApp.CleanPolicyDoc(uint(id))
	if err != nil {
		errx.Respond(c, err, "操作失败")
		return
	}
	response.OkWithDetailed(out, "清洗完成", c)
}
