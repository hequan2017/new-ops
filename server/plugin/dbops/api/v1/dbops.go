// Package api 白泽数据库工单接口
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/model/common/response"
	dbModel "github.com/hequan2017/new-ops/server/plugin/dbops/model"
	"github.com/hequan2017/new-ops/server/plugin/dbops/service"
	"github.com/hequan2017/new-ops/server/utils"
)

var DbopsApi = new(dbopsApi)

var dbSvc = new(service.DbopsService)

type dbopsApi struct{}

// CreateInstance 创建实例
// @Tags DbopsInstance
// @Summary 创建 MySQL 实例（密码 AES-GCM 加密落库）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body dbModel.DbopsInstance true "名称/主机/端口/账号/密码"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /dbops/instance [post]
func (a *dbopsApi) CreateInstance(c *gin.Context) {
	var inst dbModel.DbopsInstance
	if err := c.ShouldBindJSON(&inst); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := dbSvc.CreateInstance(&inst); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// UpdateInstance 更新实例
// @Tags DbopsInstance
// @Summary 更新实例（密码留空不改）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body dbModel.DbopsInstance true "含ID"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /dbops/instance [put]
func (a *dbopsApi) UpdateInstance(c *gin.Context) {
	var inst dbModel.DbopsInstance
	if err := c.ShouldBindJSON(&inst); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := dbSvc.UpdateInstance(&inst); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteInstance 删除实例
// @Tags DbopsInstance
// @Summary 删除实例（未结束工单拒绝）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "实例ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /dbops/instance [delete]
func (a *dbopsApi) DeleteInstance(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := dbSvc.DeleteInstance(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ListInstances 实例列表
// @Tags DbopsInstance
// @Summary 实例列表（keyword/status 过滤，密码不回显）
// @Security ApiKeyAuth
// @Produce application/json
// @Param keyword query string false "关键字"
// @Param status query string false "状态"
// @Success 200 {object} response.Response{data=[]dbModel.DbopsInstance} "获取成功"
// @Router /dbops/instance/list [get]
func (a *dbopsApi) ListInstances(c *gin.Context) {
	list, err := dbSvc.GetInstanceList(c.Query("keyword"), c.Query("status"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// TestInstance 实例连通检测
// @Tags DbopsInstance
// @Summary TCP 探活并回写状态
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "实例ID"
// @Success 200 {object} response.Response "检测完成"
// @Router /dbops/instance/test [post]
func (a *dbopsApi) TestInstance(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	online, detail, err := dbSvc.TestConnection(uint(id))
	if err != nil {
		response.FailWithMessage("检测失败（离线）: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"online": online, "detail": detail}, "检测通过", c)
}

// CreateOrder 创建 SQL 工单
// @Tags DbopsOrder
// @Summary 创建 SQL 上线工单（待审核）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "instanceId/title/sqlText"
// @Success 200 {object} response.Response{data=dbModel.DbopsOrder} "创建成功"
// @Router /dbops/order [post]
func (a *dbopsApi) CreateOrder(c *gin.Context) {
	var req struct {
		InstanceID uint   `json:"instanceId" binding:"required"`
		Title      string `json:"title" binding:"required"`
		SqlText    string `json:"sqlText" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	order, err := dbSvc.CreateOrder(req.InstanceID, req.Title, req.SqlText,
		utils.GetUserName(c), utils.GetUserID(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(order, "创建成功", c)
}

// AuditOrder 审核工单
// @Tags DbopsOrder
// @Summary SQL 审核（goInception；未配置时明确报错）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "工单ID"
// @Success 200 {object} response.Response{msg=string} "审核通过"
// @Router /dbops/order/audit [post]
func (a *dbopsApi) AuditOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if _, err := dbSvc.AuditOrder(uint(id), utils.GetUserName(c)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("审核完成", c)
}

// CancelOrder 取消工单
// @Tags DbopsOrder
// @Summary 取消待审核工单
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "工单ID"
// @Success 200 {object} response.Response{msg=string} "取消成功"
// @Router /dbops/order/cancel [post]
func (a *dbopsApi) CancelOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		response.FailWithMessage("id 无效", c)
		return
	}
	if err := dbSvc.CancelOrder(uint(id), utils.GetUserName(c)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("取消成功", c)
}

// ListOrders 工单分页
// @Tags DbopsOrder
// @Summary SQL 工单列表（普通用户仅本人）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.PageInfo true "页码/页大小"
// @Success 200 {object} response.Response{data=response.PageResult} "获取成功"
// @Router /dbops/order/list [post]
func (a *dbopsApi) ListOrders(c *gin.Context) {
	var info request.PageInfo
	_ = c.ShouldBindJSON(&info)
	list, total, err := dbSvc.GetOrderList(info.Page, info.PageSize, c.Query("creator"),
		utils.GetUserID(c), utils.GetUserAuthorityId(c) == 888)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: info.Page, PageSize: info.PageSize,
	}, "获取成功", c)
}
