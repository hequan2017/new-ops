// Package api 白泽容器管理：Compose 编排接口（M8 C2）
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/container/service"
	"github.com/hequan2017/new-ops/server/utils"
)

// composeCreateReq 创建 compose 项目请求
type composeCreateReq struct {
	Name       string `json:"name" binding:"required"`
	EndpointID uint   `json:"endpointId" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Notes      string `json:"notes"`
}

// composeActionReq compose 动作请求
type composeActionReq struct {
	ProjectID uint   `json:"projectId" binding:"required"`
	Action    string `json:"action" binding:"required"` // up|down|restart
}

// composeUpdateReq 更新 compose 内容请求
type composeUpdateReq struct {
	ProjectID uint   `json:"projectId" binding:"required"`
	Content   string `json:"content" binding:"required"`
}

// CreateComposeProject 创建 compose 项目并立即部署
// @Tags DockerCompose
// @Summary 创建 Compose 项目（校验 compose 文件后立即 up -d）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body composeCreateReq true "项目名/接入点/compose 内容"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /container/compose [post]
func (a *containerApi) CreateComposeProject(c *gin.Context) {
	var req composeCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数无效: "+err.Error(), c)
		return
	}
	p, err := service.Service.Compose.CreateComposeProject(req.Name, req.EndpointID, req.Content, utils.GetUserName(c), req.Notes)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"id": p.ID, "name": p.Name}, "创建成功", c)
}

// GetComposeProjectList compose 项目列表
// @Tags DockerCompose
// @Summary Compose 项目列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param keyword query string false "关键字"
// @Success 200 {object} response.Response{data=[]map[string]any} "获取成功"
// @Router /container/compose/list [get]
func (a *containerApi) GetComposeProjectList(c *gin.Context) {
	list, err := service.Service.Compose.GetComposeProjectList(c.Query("keyword"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// GetComposeProjectDetail compose 项目详情
// @Tags DockerCompose
// @Summary Compose 项目详情（内容 + 最近 ps 结果）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "项目ID"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /container/compose/detail [get]
func (a *containerApi) GetComposeProjectDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	d, err := service.Service.Compose.GetComposeProjectDetail(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(d, c)
}

// ComposePS compose 项目 ps
// @Tags DockerCompose
// @Summary 实时执行 docker compose ps 并回写状态
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "项目ID"
// @Success 200 {object} response.Response{data=[]service.ComposeServiceItem} "获取成功"
// @Router /container/compose/ps [get]
func (a *containerApi) ComposePS(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	items, err := service.Service.Compose.ComposePS(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(items, c)
}

// ComposeAction compose 生命周期动作
// @Tags DockerCompose
// @Summary Compose up/down/restart
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body composeActionReq true "项目ID/动作(up|down|restart)"
// @Success 200 {object} response.Response{data=string,msg=string} "执行成功"
// @Router /container/compose/action [post]
func (a *containerApi) ComposeAction(c *gin.Context) {
	var req composeActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数无效: "+err.Error(), c)
		return
	}
	out, err := service.Service.Compose.ComposeAction(req.ProjectID, req.Action)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "执行成功", c)
}

// UpdateComposeProject 更新 compose 内容
// @Tags DockerCompose
// @Summary 更新 compose 文件内容（校验后保存，配合 up 增量应用）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body composeUpdateReq true "项目ID/新内容"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /container/compose [put]
func (a *containerApi) UpdateComposeProject(c *gin.Context) {
	var req composeUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数无效: "+err.Error(), c)
		return
	}
	if err := service.Service.Compose.UpdateComposeProject(req.ProjectID, req.Content); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteComposeProject 删除 compose 项目
// @Tags DockerCompose
// @Summary 删除 Compose 项目（先 down 清理容器再删登记）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "项目ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /container/compose [delete]
func (a *containerApi) DeleteComposeProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := service.Service.Compose.DeleteComposeProject(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
