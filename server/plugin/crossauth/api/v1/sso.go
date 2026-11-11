// Package v1 crossauth API：SSO 单点登录（需求 3.7-4）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// issueTicketReq 签发请求
type issueTicketReq struct {
	AppKey   string `json:"appKey" binding:"required"`
	UserName string `json:"userName" binding:"required"`
}

// IssueCaSsoTicket 签发票据
// @Tags Crossauth
// @Summary SSO 一次性票据签发（60s 有效，返回跳转 URL）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.issueTicketReq true "签发请求"
// @Success 200 {object} response.Response{data=map[string]any} "签发成功"
// @Router /crossauth/sso/ticket [post]
func (a *CrossauthApi) IssueCaSsoTicket(c *gin.Context) {
	var in issueTicketReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	ticket, redirect, err := service.CaApp.IssueCaSsoTicket(in.AppKey, in.UserName)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"ticket": ticket.Ticket, "expiredAt": ticket.ExpiredAt, "redirectUrl": redirect}, "签发成功", c)
}

// redeemReq 换取请求
type redeemReq struct {
	Ticket string `json:"ticket" binding:"required"`
	AppKey string `json:"appKey"`
}

// RedeemCaSsoTicket 换取票据
// @Tags Crossauth
// @Summary SSO 票据换取（防重放：已用/过期/应用不符拒绝）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.redeemReq true "换取请求"
// @Success 200 {object} response.Response{data=model.CaSsoTicket} "换取成功"
// @Router /crossauth/sso/redeem [post]
func (a *CrossauthApi) RedeemCaSsoTicket(c *gin.Context) {
	var in redeemReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.RedeemCaSsoTicket(in.Ticket, in.AppKey)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "换取成功", c)
}

// SaveCaSsoApp 注册应用
// @Tags Crossauth
// @Summary 注册/更新 SSO 应用（回调地址校验）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.CaSsoApp true "应用"
// @Success 200 {object} response.Response{data=model.CaSsoApp} "保存成功"
// @Router /crossauth/sso/app [post]
func (a *CrossauthApi) SaveCaSsoApp(c *gin.Context) {
	var in model.CaSsoApp
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.SaveCaSsoApp(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetCaSsoApps 应用列表
// @Tags Crossauth
// @Summary SSO 应用列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.CaSsoApp} "获取成功"
// @Router /crossauth/sso/apps [get]
func (a *CrossauthApi) GetCaSsoApps(c *gin.Context) {
	list, err := service.CaApp.GetCaSsoApps()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// CleanupCaSsoTickets 票据清理
// @Tags Crossauth
// @Summary 过期/已用票据清理（已用默认保留 24h 供审计，参数可调）
// @Security ApiKeyAuth
// @Produce application/json
// @Param keepHours query int false "已用票据保留小时数(默认 24)"
// @Success 200 {object} response.Response{data=map[string]any} "清理完成"
// @Router /crossauth/sso/cleanup [post]
func (a *CrossauthApi) CleanupCaSsoTickets(c *gin.Context) {
	keepHours, _ := strconv.Atoi(c.DefaultQuery("keepHours", "24"))
	out, err := service.CaApp.CleanupCaSsoTickets(keepHours)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "清理完成", c)
}

// GetCaSsoTicketStats 票据统计
// @Tags Crossauth
// @Summary SSO 票据存量统计（总量/有效/已用/过期）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=service.CaSsoTicketStats} "获取成功"
// @Router /crossauth/sso/stats [get]
func (a *CrossauthApi) GetCaSsoTicketStats(c *gin.Context) {
	out, err := service.CaApp.GetCaSsoTicketStats()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "获取成功", c)
}
