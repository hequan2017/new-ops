// Package api 流水线构建接口
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/pipeline/model"
	"github.com/hequan2017/new-ops/server/utils"
)

type pipelineBuild struct{}

// PipelineBuildApi 构建接口实例（路由组注册用）
var PipelineBuildApi = new(pipelineBuild)

// StartBuild 触发构建
// @Tags PipelineBuild
// @Summary 触发流水线构建（异步执行，返回构建记录）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "pipelineId/params(map[string]string，渲染步骤 {{key}})"
// @Success 200 {object} response.Response{data=model.PipelineBuild} "已触发"
// @Router /pipeline/build/start [post]
func (p *pipelineBuild) StartBuild(c *gin.Context) {
	var req struct {
		PipelineID uint              `json:"pipelineId" binding:"required"`
		Params     map[string]string `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	build, err := buildSvc.CreateBuild(req.PipelineID, req.Params, utils.GetUserName(c), utils.GetUserID(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(build, "构建已触发", c)
}

// CancelBuild 取消构建
// @Tags PipelineBuild
// @Summary 取消构建（等待审批/执行中）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "构建ID"
// @Success 200 {object} response.Response "取消成功"
// @Router /pipeline/build/cancel [post]
func (p *pipelineBuild) CancelBuild(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := buildSvc.CancelBuild(uint(id), utils.GetUserName(c)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("已取消", c)
}

// ApproveBuild 审批放行
// @Tags PipelineBuild
// @Summary 人工审批放行（等待审批状态的构建）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "构建ID"
// @Success 200 {object} response.Response "已放行"
// @Router /pipeline/build/approve [post]
func (p *pipelineBuild) ApproveBuild(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := buildSvc.ApproveBuild(uint(id), utils.GetUserName(c)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("已放行", c)
}

// GetBuildList 构建分页列表
// @Tags PipelineBuild
// @Summary 构建分页列表（pipelineId/status 过滤，普通用户仅本人）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.PageInfo true "页码/页大小"
// @Param pipelineId query int false "流水线ID"
// @Param status query string false "状态"
// @Success 200 {object} response.Response{data=response.PageResult} "获取成功"
// @Router /pipeline/build/list [post]
func (p *pipelineBuild) GetBuildList(c *gin.Context) {
	var info struct {
		Page       int `json:"page"`
		PageSize   int `json:"pageSize"`
		PipelineID uint `json:"pipelineId"`
	}
	_ = c.ShouldBindJSON(&info)
	var pipelineID *uint
	if info.PipelineID > 0 {
		pipelineID = &info.PipelineID
	}
	list, total, err := buildSvc.GetBuildList(info.Page, info.PageSize, pipelineID,
		c.Query("status"), c.Query("username"), utils.GetUserID(c), utils.GetUserAuthorityId(c) == 888)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetBuildLogs 构建日志
// @Tags PipelineBuild
// @Summary 构建日志（升序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "构建ID"
// @Param page query int false "页码"
// @Param pageSize query int false "页大小(默认100)"
// @Success 200 {object} response.Response{data=response.PageResult} "获取成功"
// @Router /pipeline/build/logs [get]
func (p *pipelineBuild) GetBuildLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	list, total, err := buildSvc.GetBuildLogs(uint(id), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// StreamBuildLogs 构建日志 SSE 实时流
// @Tags PipelineBuild
// @Summary 构建日志 SSE（增量推送，构建结束自动收流）
// @Security ApiKeyAuth
// @Param token query string true "JWT"
// @Param id query int true "构建ID"
// @Param lastId query int false "起始日志ID（断线重连续传）"
// @Success 200 {string} string "text/event-stream"
// @Router /sse/pipeline/build/logs [get]
func (p *pipelineBuild) StreamBuildLogs(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.String(http.StatusUnauthorized, "缺少 token")
		return
	}
	if _, err := utils.NewJWT().ParseToken(token); err != nil {
		c.String(http.StatusUnauthorized, "token 无效: "+err.Error())
		return
	}
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		c.String(http.StatusBadRequest, "id 无效")
		return
	}
	lastID, _ := strconv.ParseUint(c.DefaultQuery("lastId", "0"), 10, 64)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.String(http.StatusInternalServerError, "不支持流式响应")
		return
	}
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	writeEvent := func(event string, payload any) bool {
		data, jErr := json.Marshal(payload)
		if jErr != nil {
			return true
		}
		if _, wErr := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data); wErr != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	// 先回当前状态，随后增量推日志；构建结束且无新日志后发 done 收流
	status, _ := buildSvc.GetBuildStatus(uint(id))
	_ = writeEvent("status", gin.H{"status": status})
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		logs, lErr := buildSvc.GetBuildLogsAfter(uint(id), uint(lastID))
		if lErr == nil {
			for _, l := range logs {
				if !writeEvent("log", l) {
					return // 客户端断开
				}
				lastID = uint64(l.ID)
			}
		}
		status, _ = buildSvc.GetBuildStatus(uint(id))
		finished := status == model.BuildSuccess || status == model.BuildFailed || status == model.BuildCanceled
		if finished {
			_ = writeEvent("status", gin.H{"status": status})
			_ = writeEvent("done", gin.H{"buildId": id, "status": status})
			return
		}
		select {
		case <-c.Request.Context().Done():
			return
		case <-tick.C:
		}
	}
}
