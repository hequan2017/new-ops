// Package api 白泽容器管理接口层
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
	ctModel "github.com/hequan2017/new-ops/server/plugin/container/model"
	"github.com/hequan2017/new-ops/server/plugin/container/service"
)

var ContainerApi = new(containerApi)

// ctSvc 容器实时服务实例
var ctSvc = &service.Service.Endpoint

type containerApi struct{}

// CreateEndpoint 创建接入点
// @Tags DockerEndpoint
// @Summary 创建 Docker 接入点
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body ctModel.DockerEndpoint true "名称/地址(unix://|tcp://)/TLS凭据"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /container/endpoint [post]
func (a *containerApi) CreateEndpoint(c *gin.Context) {
	var ep ctModel.DockerEndpoint
	if err := c.ShouldBindJSON(&ep); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.Service.Endpoint.CreateEndpoint(&ep); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateEndpoint 更新接入点
// @Tags DockerEndpoint
// @Summary 更新 Docker 接入点
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body ctModel.DockerEndpoint true "含ID"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /container/endpoint [put]
func (a *containerApi) UpdateEndpoint(c *gin.Context) {
	var ep ctModel.DockerEndpoint
	if err := c.ShouldBindJSON(&ep); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.Service.Endpoint.UpdateEndpoint(&ep); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteEndpoint 删除接入点
// @Tags DockerEndpoint
// @Summary 删除 Docker 接入点
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "接入点ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /container/endpoint [delete]
func (a *containerApi) DeleteEndpoint(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := service.Service.Endpoint.DeleteEndpoint(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetEndpointList 接入点列表
// @Tags DockerEndpoint
// @Summary 接入点列表（keyword/status 过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param keyword query string false "关键字"
// @Param status query string false "状态"
// @Success 200 {object} response.Response{data=[]ctModel.DockerEndpoint} "获取成功"
// @Router /container/endpoint/list [get]
func (a *containerApi) GetEndpointList(c *gin.Context) {
	list, err := service.Service.Endpoint.GetEndpointList(c.Query("keyword"), c.Query("status"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// CheckEndpoint 手动巡检
// @Tags DockerEndpoint
// @Summary 手动连通巡检（Ping+版本回写）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "接入点ID"
// @Success 200 {object} response.Response "巡检完成"
// @Router /container/endpoint/check [post]
func (a *containerApi) CheckEndpoint(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	ep, err := service.Service.Endpoint.GetEndpoint(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	online, version, pingErr := service.Service.Endpoint.PingEndpoint(ep)
	if pingErr != nil {
		response.FailWithMessage("巡检失败（离线）: "+pingErr.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"online": online, "version": version}, "巡检通过", c)
}
