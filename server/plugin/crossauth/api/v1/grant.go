// Package v1 crossauth API：权限分配双模式（需求 3.7-6）
package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// SaveCaAutoGrantRule 保存自动匹配规则
// @Tags Crossauth
// @Summary 保存院校/部门→角色自动匹配规则（dept=* 通配）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.CaAutoGrantRule true "规则"
// @Success 200 {object} response.Response{data=model.CaAutoGrantRule} "保存成功"
// @Router /crossauth/grant/rule [post]
func (a *CrossauthApi) SaveCaAutoGrantRule(c *gin.Context) {
	var in model.CaAutoGrantRule
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.SaveCaAutoGrantRule(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// autoGrantReq 自动匹配请求
type autoGrantReq struct {
	UserName string `json:"userName" binding:"required"`
	School   string `json:"school" binding:"required"`
	Dept     string `json:"dept"`
}

// AutoGrantCa 自动匹配分配
// @Tags Crossauth
// @Summary 自动匹配分配（院校+部门→模板→生效权限，无匹配提示走人工）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.autoGrantReq true "匹配请求"
// @Success 200 {object} response.Response{data=model.CaGrantAssignment} "分配成功"
// @Router /crossauth/grant/auto [post]
func (a *CrossauthApi) AutoGrantCa(c *gin.Context) {
	var in autoGrantReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.AutoGrantCa(in.UserName, in.School, in.Dept)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "分配成功", c)
}

// SubmitCaGrantRequest 提交申请
// @Tags Crossauth
// @Summary 提交人工审核申请（本校+归属校双重审核）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.CaGrantRequest true "申请"
// @Success 200 {object} response.Response{data=model.CaGrantRequest} "提交成功"
// @Router /crossauth/grant/request [post]
func (a *CrossauthApi) SubmitCaGrantRequest(c *gin.Context) {
	var in model.CaGrantRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.SubmitCaGrantRequest(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "提交成功", c)
}

// approveReq 审核请求
type approveReq struct {
	ID       uint   `json:"id" binding:"required"`
	Approver string `json:"approver" binding:"required"`
}

// ApproveCaGrantRequest 审核推进
// @Tags Crossauth
// @Summary 审核推进（提交→本校审核→归属校审核→通过，通过即建分配+实时推送留痕）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.approveReq true "审核请求"
// @Success 200 {object} response.Response{data=model.CaGrantRequest} "推进成功"
// @Router /crossauth/grant/approve [post]
func (a *CrossauthApi) ApproveCaGrantRequest(c *gin.Context) {
	var in approveReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.ApproveCaGrantRequest(in.ID, in.Approver)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "推进成功", c)
}

// RejectCaGrantRequest 驳回
// @Tags Crossauth
// @Summary 申请驳回（任一环节，实时推送留痕）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.approveReq true "驳回请求"
// @Success 200 {object} response.Response{data=model.CaGrantRequest} "驳回成功"
// @Router /crossauth/grant/reject [post]
func (a *CrossauthApi) RejectCaGrantRequest(c *gin.Context) {
	var in approveReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.RejectCaGrantRequest(in.ID, in.Approver)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "驳回成功", c)
}

// GetCaGrantRequests 申请列表
// @Tags Crossauth
// @Summary 审核申请列表（状态过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param status query string false "提交/本校审核/归属校审核/通过/驳回"
// @Success 200 {object} response.Response{data=[]model.CaGrantRequest} "获取成功"
// @Router /crossauth/grant/requests [get]
func (a *CrossauthApi) GetCaGrantRequests(c *gin.Context) {
	list, err := service.CaApp.GetCaGrantRequests(c.Query("status"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// GetCaGrantAssignments 分配列表
// @Tags Crossauth
// @Summary 权限分配列表（用户过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param userName query string false "用户"
// @Success 200 {object} response.Response{data=[]model.CaGrantAssignment} "获取成功"
// @Router /crossauth/grant/assignments [get]
func (a *CrossauthApi) GetCaGrantAssignments(c *gin.Context) {
	list, err := service.CaApp.GetCaGrantAssignments(c.Query("userName"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
