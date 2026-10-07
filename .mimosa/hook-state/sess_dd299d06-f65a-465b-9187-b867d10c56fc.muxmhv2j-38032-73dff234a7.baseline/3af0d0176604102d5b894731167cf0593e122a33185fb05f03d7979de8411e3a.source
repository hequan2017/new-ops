// Package api 终端会话审计查询接口（仅 888）
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/model/common/response"
	service "github.com/hequan2017/new-ops/server/plugin/term/service"
)

type termSession struct{}

// GetTermSessionList 会话分页列表
// @Tags TermSession
// @Summary 会话分页列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param data body request.PageInfo true "页码/页大小"
// @Param username query string false "操作用户"
// @Param status query string false "状态(进行中/已结束)"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /term/session/list [post]
func (t *termSession) GetTermSessionList(c *gin.Context) {
	var info request.PageInfo
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := service.Service.TermAuditService.GetSessionList(info, c.Query("username"), c.Query("status"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: info.Page, PageSize: info.PageSize,
	}, "获取成功", c)
}

// GetTermSessionStreams 会话流镜像（回放数据）
// @Tags TermSession
// @Summary 会话流镜像
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "会话ID"
// @Success 200 {object} response.Response{data=[]model.TermSessionStream} "获取成功"
// @Router /term/session/streams [get]
func (t *termSession) GetTermSessionStreams(c *gin.Context) {
	id, err := parseUintQuery(c.Query("id"))
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	list, err := service.Service.TermAuditService.GetSessionStreams(id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// GetTermSessionCommands 会话命令列表
// @Tags TermSession
// @Summary 会话命令列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "会话ID"
// @Success 200 {object} response.Response{data=[]model.TermSessionCommand} "获取成功"
// @Router /term/session/commands [get]
func (t *termSession) GetTermSessionCommands(c *gin.Context) {
	id, err := parseUintQuery(c.Query("id"))
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	list, err := service.Service.TermAuditService.GetSessionCommands(id)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}
