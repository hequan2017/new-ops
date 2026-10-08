// Package api 白泽容器管理：端口转发规则接口（M8 C2）
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/container/service"
)

// portRuleSetReq 应用端口规则请求
type portRuleSetReq struct {
	EndpointID  uint               `json:"endpointId" binding:"required"`
	ContainerID string             `json:"containerId" binding:"required"`
	Rules       []service.PortRule `json:"rules" binding:"required"`
}

// ListPortForwards 容器端口映射列表
// @Tags ContainerPortForward
// @Summary 读取容器当前端口转发规则
// @Security ApiKeyAuth
// @Produce application/json
// @Param endpointId query int true "接入点ID"
// @Param id query string true "容器ID"
// @Success 200 {object} response.Response{data=[]service.PortRule} "获取成功"
// @Router /container/container/portforwards [get]
func (a *containerApi) ListPortForwards(c *gin.Context) {
	endpointID, err := strconv.ParseUint(c.Query("endpointId"), 10, 64)
	if err != nil || endpointID == 0 {
		response.FailWithMessage("endpointId 无效", c)
		return
	}
	cid := c.Query("id")
	if cid == "" {
		response.FailWithMessage("容器ID必填", c)
		return
	}
	rules, err := ctSvc.ListPortForwards(uint(endpointID), cid)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(rules, c)
}

// SetPortForwards 应用端口转发规则
// @Tags ContainerPortForward
// @Summary 应用端口转发规则（Docker 端口绑定不可在线修改，将按新规则重建容器；写操作仅 888）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body portRuleSetReq true "接入点/容器/规则列表"
// @Success 200 {object} response.Response{data=string,msg=string} "应用成功"
// @Router /container/container/portforwards [post]
func (a *containerApi) SetPortForwards(c *gin.Context) {
	var req portRuleSetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数无效: "+err.Error(), c)
		return
	}
	newID, err := ctSvc.SetPortForwards(req.EndpointID, req.ContainerID, req.Rules)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(newID, "端口规则已应用（容器已按新规则重建）", c)
}
