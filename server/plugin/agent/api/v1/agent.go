// Package api 白泽 Agent 通道接口层（M9 A1）
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
	agentSvc "github.com/hequan2017/new-ops/server/plugin/agent/service"
)

var AgentApi = new(agentApi)

type agentApi struct{}

// tokenIssueReq 令牌签发请求
type tokenIssueReq struct {
	AssetID uint   `json:"assetId" binding:"required"`
	Labels  string `json:"labels"`
}

// IssueToken 签发/重置 Agent 接入令牌
// @Tags AgentInstance
// @Summary 为资产签发/重置 Agent 接入令牌（明文仅本次返回，落库仅 SHA256）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body tokenIssueReq true "资产ID/标签"
// @Success 200 {object} response.Response{data=string,msg=string} "签发成功"
// @Router /agent/instance/token [post]
func (a *agentApi) IssueToken(c *gin.Context) {
	var req tokenIssueReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数无效: "+err.Error(), c)
		return
	}
	inst, token, err := agentSvc.AgentSvc.IssueAgentToken(req.AssetID, req.Labels)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{
		"id": inst.ID, "assetId": inst.AssetID, "token": token,
		"tokenHint": inst.TokenHint,
	}, "签发成功（明文仅此一次返回，请立即保存）", c)
}

// ListInstances Agent 实例列表
// @Tags AgentInstance
// @Summary Agent 实例列表（状态按心跳时效计算：90s 无心跳离线）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]service.AgentInstanceView} "获取成功"
// @Router /agent/instance/list [get]
func (a *agentApi) ListInstances(c *gin.Context) {
	list, err := agentSvc.AgentSvc.ListAgentInstances()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// RevokeInstance 吊销 Agent 注册
// @Tags AgentInstance
// @Summary 吊销 Agent 注册（删除登记，Agent 下次注册被拒；写操作仅 888）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "实例ID"
// @Success 200 {object} response.Response{msg=string} "吊销成功"
// @Router /agent/instance [delete]
func (a *agentApi) RevokeInstance(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := agentSvc.AgentSvc.RevokeAgentToken(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("吊销成功", c)
}
