// Package v1 eduadmin API：自动报销（需求 3.6-16）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SubmitEduExpense 提交报销
// @Tags Eduadmin
// @Summary 提交报销单（登记+规则自动审核：限额/类别/重复报销）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduInvoice true "报销单"
// @Success 200 {object} response.Response{data=model.EduInvoice} "提交成功"
// @Router /eduadmin/expense [post]
func (a *EduadminApi) SubmitEduExpense(c *gin.Context) {
	var in model.EduInvoice
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SubmitEduExpense(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "提交成功", c)
}

// GetEduInvoices 报销列表
// @Tags Eduadmin
// @Summary 报销单列表（状态/部门过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param status query string false "提交/审核中/支付/驳回"
// @Param dept query string false "部门"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/expense/list [get]
func (a *EduadminApi) GetEduInvoices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.EduadminApp.GetEduInvoices(c.Query("status"), c.Query("dept"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// PayEduExpense 支付
// @Tags Eduadmin
// @Summary 报销支付（审核中→支付）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "报销单ID"
// @Success 200 {object} response.Response{data=model.EduInvoice} "支付成功"
// @Router /eduadmin/expense/pay [post]
func (a *EduadminApi) PayEduExpense(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	out, err := service.EduadminApp.PayEduExpense(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "支付成功", c)
}

// RejectEduExpense 人工驳回
// @Tags Eduadmin
// @Summary 报销人工驳回（审核中→驳回）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.rejectExpenseReq true "驳回请求"
// @Success 200 {object} response.Response{data=model.EduInvoice} "驳回成功"
// @Router /eduadmin/expense/reject [post]
func (a *EduadminApi) RejectEduExpense(c *gin.Context) {
	var in struct {
		ID     uint   `json:"id" binding:"required"`
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.RejectEduExpense(in.ID, in.Reason)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "驳回成功", c)
}
