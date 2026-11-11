// Package v1 eduadmin API：政策数据更新（需求 3.6-35）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// PullEduGovPolicy 拉取政策
// @Tags Eduadmin
// @Summary 拉取上级政策条目（教育部/上海市教委，SHA256(title|url) 幂等入库+时效性字段）
// @Security ApiKeyAuth
// @Produce application/json
// @Param source query string true "教育部/上海市教委"
// @Param endpoint query string true "拉取端点"
// @Success 200 {object} response.Response{data=map[string]any} "拉取完成"
// @Router /eduadmin/govPolicy/pull [post]
func (a *EduadminApi) PullEduGovPolicy(c *gin.Context) {
	out, err := service.EduadminApp.PullEduGovPolicy(c.Query("source"), c.Query("endpoint"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "拉取完成", c)
}

// GetEduGovPolicies 政策列表
// @Tags Eduadmin
// @Summary 上级政策条目列表（来源过滤，发布时间倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param source query string false "教育部/上海市教委"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/govPolicy/list [get]
func (a *EduadminApi) GetEduGovPolicies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.EduadminApp.GetEduGovPolicies(c.Query("source"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}
