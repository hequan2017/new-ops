// Package v1 artsci API：文理学科 AI 工具集接口层（参数绑定与响应，禁止业务逻辑）
package v1

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/artsci/model"
	"github.com/hequan2017/new-ops/server/plugin/artsci/service"
)

var Api = new(ArtsciApi)

type ArtsciApi struct{}

// runToolReq 执行工具请求
type runToolReq struct {
	ID       uint   `json:"id" binding:"required"`
	ArgsJSON string `json:"argsJSON"`
}

// Ping 插件连通性探测
// @Tags Artsci
// @Summary artsci 连通性探测
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]any} "探测成功"
// @Router /artsci/ping [get]
func (a *ArtsciApi) Ping(c *gin.Context) {
	response.OkWithDetailed(gin.H{
		"plugin": "artsci",
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}, "探测成功", c)
}

// GetTools 工具列表
// @Tags Artsci
// @Summary 学科工具列表（学科/关键字过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param subject query string false "学科(数学/文科/图像)"
// @Param keyword query string false "名称/描述关键字"
// @Success 200 {object} response.Response{data=[]model.ArtsciTool} "获取成功"
// @Router /artsci/tool/list [get]
func (a *ArtsciApi) GetTools(c *gin.Context) {
	list, err := service.ArtsciApp.GetTools(c.Query("subject"), c.Query("keyword"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveTool 保存工具
// @Tags Artsci
// @Summary 新建/更新学科工具
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.ArtsciTool true "工具"
// @Success 200 {object} response.Response{data=model.ArtsciTool} "保存成功"
// @Router /artsci/tool [post]
func (a *ArtsciApi) SaveTool(c *gin.Context) {
	var in model.ArtsciTool
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.ArtsciApp.SaveTool(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeleteTool 删除工具
// @Tags Artsci
// @Summary 删除学科工具（软删除）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "工具ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /artsci/tool [delete]
func (a *ArtsciApi) DeleteTool(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.ArtsciApp.DeleteTool(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// RunTool 执行工具
// @Tags Artsci
// @Summary 执行学科工具（module 分发内置实现，成败均留痕）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.runToolReq true "执行请求"
// @Success 200 {object} response.Response{data=model.ArtsciRunLog} "执行成功"
// @Router /artsci/tool/run [post]
func (a *ArtsciApi) RunTool(c *gin.Context) {
	var in runToolReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	log, err := service.ArtsciApp.RunTool(in.ID, in.ArgsJSON)
	if err != nil {
		if log != nil {
			response.OkWithDetailed(log, "执行失败（已留痕）: "+err.Error(), c)
			return
		}
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(log, "执行成功", c)
}

// GetRunLogs 执行留痕
// @Tags Artsci
// @Summary 工具执行留痕列表（工具过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param toolId query int false "工具ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /artsci/tool/runLogs [get]
func (a *ArtsciApi) GetRunLogs(c *gin.Context) {
	toolId, _ := strconv.ParseUint(c.Query("toolId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ArtsciApp.GetRunLogs(uint(toolId), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}
