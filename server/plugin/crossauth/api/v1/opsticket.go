// Package v1 crossauth API：跨校联合运维（需求 3.7-15）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// createOpsReq 工单创建请求
type createOpsReq struct {
	School string `json:"school" binding:"required"`
	Title  string `json:"title" binding:"required"`
}

// CreateCaOpsTicket 创建工单
// @Tags Crossauth
// @Summary 创建联合运维工单（提交态）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.createOpsReq true "工单"
// @Success 200 {object} response.Response{data=map[string]any} "创建成功"
// @Router /crossauth/opsTicket [post]
func (a *CrossauthApi) CreateCaOpsTicket(c *gin.Context) {
	var in createOpsReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.CreateCaOpsTicket(in.School, in.Title)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "创建成功", c)
}

// GetCaOpsTickets 工单列表
// @Tags Crossauth
// @Summary 运维工单列表（状态/院校过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param status query string false "提交/分配/处理/闭环"
// @Param school query string false "院校"
// @Success 200 {object} response.Response{data=[]model.CaOpsTicket} "获取成功"
// @Router /crossauth/opsTicket/list [get]
func (a *CrossauthApi) GetCaOpsTickets(c *gin.Context) {
	list, err := service.CaApp.GetCaOpsTickets(c.Query("status"), c.Query("school"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// AdvanceCaOpsTicket 推进
// @Tags Crossauth
// @Summary 工单推进（提交→分配→处理→闭环；分配须承接方）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "工单ID"
// @Param assignee query string false "承接运维方"
// @Success 200 {object} response.Response{data=map[string]any} "推进成功"
// @Router /crossauth/opsTicket/advance [post]
func (a *CrossauthApi) AdvanceCaOpsTicket(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	out, err := service.CaApp.AdvanceCaOpsTicket(uint(id), c.Query("assignee"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "推进成功", c)
}

// messageReq 消息请求
type messageReq struct {
	TicketId uint   `json:"ticketId" binding:"required"`
	From     string `json:"from" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

// PostCaOpsMessage 发消息
// @Tags Crossauth
// @Summary 工单协同消息（文字口径）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.messageReq true "消息"
// @Success 200 {object} response.Response{data=map[string]any} "发送成功"
// @Router /crossauth/opsTicket/message [post]
func (a *CrossauthApi) PostCaOpsMessage(c *gin.Context) {
	var in messageReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.PostCaOpsMessage(in.TicketId, in.From, in.Content)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "发送成功", c)
}

// GetCaOpsMessages 消息列表
// @Tags Crossauth
// @Summary 工单协同消息列表（升序）
// @Security ApiKeyAuth
// @Produce application/json
// @Param ticketId query int true "工单ID"
// @Success 200 {object} response.Response{data=[]model.CaOpsMessage} "获取成功"
// @Router /crossauth/opsTicket/messages [get]
func (a *CrossauthApi) GetCaOpsMessages(c *gin.Context) {
	ticketId, _ := strconv.ParseUint(c.Query("ticketId"), 10, 64)
	list, err := service.CaApp.GetCaOpsMessages(uint(ticketId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// assistReq 授权请求
type assistReq struct {
	TicketId  uint   `json:"ticketId" binding:"required"`
	Requester string `json:"requester" binding:"required"`
	Approver  string `json:"approver" binding:"required"`
	Scope     string `json:"scope" binding:"required"`
}

// GrantCaRemoteAssist 远程协助授权
// @Tags Crossauth
// @Summary 远程协助授权（范围 只读/操作级）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.assistReq true "授权"
// @Success 200 {object} response.Response{data=map[string]any} "授权成功"
// @Router /crossauth/remoteAssist [post]
func (a *CrossauthApi) GrantCaRemoteAssist(c *gin.Context) {
	var in assistReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.GrantCaRemoteAssist(in.TicketId, in.Requester, in.Approver, in.Scope)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "授权成功", c)
}

// RevokeCaRemoteAssist 撤销授权
// @Tags Crossauth
// @Summary 撤销远程协助授权
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "授权ID"
// @Success 200 {object} response.Response "撤销成功"
// @Router /crossauth/remoteAssist/revoke [post]
func (a *CrossauthApi) RevokeCaRemoteAssist(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if err := service.CaApp.RevokeCaRemoteAssist(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("撤销成功", c)
}

// GetCaRemoteAssists 授权记录
// @Tags Crossauth
// @Summary 远程协助授权记录（工单过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param ticketId query int false "工单ID"
// @Success 200 {object} response.Response{data=[]model.CaRemoteAssist} "获取成功"
// @Router /crossauth/remoteAssist/list [get]
func (a *CrossauthApi) GetCaRemoteAssists(c *gin.Context) {
	ticketId, _ := strconv.ParseUint(c.Query("ticketId"), 10, 64)
	list, err := service.CaApp.GetCaRemoteAssists(uint(ticketId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
