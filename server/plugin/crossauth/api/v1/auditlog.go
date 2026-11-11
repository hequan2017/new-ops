// Package v1 crossauth API：三级审计（需求 3.7-14）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// RecordCaAuditEvent 登记事件
// @Tags Crossauth
// @Summary 登记审计事件（校级/院系级/用户级）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.CaAuditEvent true "事件"
// @Success 200 {object} response.Response{data=model.CaAuditEvent} "登记成功"
// @Router /crossauth/auditEvent [post]
func (a *CrossauthApi) RecordCaAuditEvent(c *gin.Context) {
	var in model.CaAuditEvent
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.RecordCaAuditEvent(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "登记成功", c)
}

// GetCaAuditEvents 分层查询
// @Tags Crossauth
// @Summary 三级审计分层查询（校级全量可筛校/院系级须校+院系/用户级须用户）
// @Security ApiKeyAuth
// @Produce application/json
// @Param level query string true "校级/院系级/用户级"
// @Param school query string false "院校"
// @Param dept query string false "院系"
// @Param userName query string false "用户"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /crossauth/auditEvent/list [get]
func (a *CrossauthApi) GetCaAuditEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	list, total, err := service.CaApp.GetCaAuditEvents(c.Query("level"), c.Query("school"), c.Query("dept"), c.Query("userName"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}
