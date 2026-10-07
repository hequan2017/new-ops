// Package api 白泽批量作业接口层
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/model/common/response"
	jobRequest "github.com/hequan2017/new-ops/server/plugin/job/model/request"
	"github.com/hequan2017/new-ops/server/plugin/job/service"
	"github.com/hequan2017/new-ops/server/utils"
)

type batchExec struct{}

// CreateBatchExec 发起批量命令执行
// @Tags Job
// @Summary 发起批量命令执行（异步，返回批次）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.BatchExecReq true "主机列表/命令/并发/超时/统一凭据"
// @Success 200 {object} response.Response{data=model.JobExecRecord} "创建成功"
// @Router /job/exec [post]
func (b *batchExec) CreateBatchExec(c *gin.Context) {
	var req jobRequest.BatchExecReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	record, err := service.JobService.CreateBatchExec(req, utils.GetUserID(c), utils.GetUserAuthorityId(c) == 888, utils.GetUserName(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(record, "批次已创建，正在后台执行", c)
}

// CancelBatchExec 取消进行中的批次
// @Tags Job
// @Summary 取消进行中的批次
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "批次ID"
// @Success 200 {object} response.Response "取消成功"
// @Router /job/exec/cancel [post]
func (b *batchExec) CancelBatchExec(c *gin.Context) {
	id, err := parseUintQuery(c.Query("id"))
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := service.JobService.CancelBatchExec(id); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("已发送取消信号", c)
}

// GetBatchList 批次分页列表
// @Tags Job
// @Summary 批次分页列表（普通用户仅见自己发起的批次）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.PageInfo true "页码/页大小"
// @Param status query string false "状态过滤(执行中/已完成/已取消)"
// @Success 200 {object} response.Response{data=response.PageResult} "获取成功"
// @Router /job/exec/list [post]
func (b *batchExec) GetBatchList(c *gin.Context) {
	var info request.PageInfo
	if err := c.ShouldBindJSON(&info); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := service.JobService.GetRecordList(info, utils.GetUserID(c), utils.GetUserAuthorityId(c) == 888, c.Query("status"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: info.Page, PageSize: info.PageSize,
	}, "获取成功", c)
}

// GetBatchDetail 批次详情（含每主机结果）
// @Tags Job
// @Summary 批次详情与每主机执行结果
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "批次ID"
// @Success 200 {object} response.Response{data=model.JobExecRecord} "获取成功（results 字段附带每主机结果）"
// @Router /job/exec/detail [get]
func (b *batchExec) GetBatchDetail(c *gin.Context) {
	id, err := parseUintQuery(c.Query("id"))
	if err != nil || id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	record, results, err := service.JobService.GetRecordDetail(id, utils.GetUserID(c), utils.GetUserAuthorityId(c) == 888)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"record": record, "results": results}, "获取成功", c)
}

// parseUintQuery 安全解析 uint 查询参数
func parseUintQuery(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	return uint(v), err
}
