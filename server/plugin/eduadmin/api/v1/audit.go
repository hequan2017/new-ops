// Package v1 eduadmin API：审计知识语料（需求 3.6-37）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SaveEduAuditItem 保存条目
// @Tags Eduadmin
// @Summary 新建/更新审计知识条目（六域白名单）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduAuditItem true "条目"
// @Success 200 {object} response.Response{data=model.EduAuditItem} "保存成功"
// @Router /eduadmin/audit [post]
func (a *EduadminApi) SaveEduAuditItem(c *gin.Context) {
	var in model.EduAuditItem
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduAuditItem(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetEduAuditItems 条目列表
// @Tags Eduadmin
// @Summary 审计知识条目列表（域过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param domain query string false "财务/采购/人事/资产/工程/数据安全"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/audit/list [get]
func (a *EduadminApi) GetEduAuditItems(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.EduadminApp.GetEduAuditItems(c.Query("domain"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// SearchEduAudit 审计检索
// @Tags Eduadmin
// @Summary 审计知识 bigram 检索（辅助审计与内控信息化）
// @Security ApiKeyAuth
// @Produce application/json
// @Param query query string true "检索词"
// @Param domain query string false "域"
// @Param topK query int false "返回条数"
// @Success 200 {object} response.Response{data=[]service.auditHit} "检索成功"
// @Router /eduadmin/audit/search [get]
func (a *EduadminApi) SearchEduAudit(c *gin.Context) {
	topK, _ := strconv.Atoi(c.DefaultQuery("topK", "5"))
	hits, err := service.EduadminApp.SearchEduAudit(c.Query("query"), c.Query("domain"), topK)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(hits, "检索成功", c)
}
