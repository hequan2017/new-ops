// Package v1 eduadmin API：设施维护（需求 3.6-22）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// createRepairReq 工单创建请求
type createRepairReq struct {
	AssetId     uint   `json:"assetId"`
	Description string `json:"description" binding:"required"`
}

// CreateEduRepairOrder 创建工单
// @Tags Eduadmin
// @Summary 创建维修工单（故障预测联动自动派发：等级路由班组）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.createRepairReq true "工单请求"
// @Success 200 {object} response.Response{data=model.EduRepairOrder} "创建成功"
// @Router /eduadmin/repair [post]
func (a *EduadminApi) CreateEduRepairOrder(c *gin.Context) {
	var in createRepairReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.CreateEduRepairOrder(in.AssetId, in.Description)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "创建成功", c)
}

// GetEduRepairOrders 工单列表
// @Tags Eduadmin
// @Summary 维修工单列表（状态过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param status query string false "待派/处理中/已闭环"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/repair/list [get]
func (a *EduadminApi) GetEduRepairOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.EduadminApp.GetEduRepairOrders(c.Query("status"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// assignRepairReq 派单请求
type assignRepairReq struct {
	ID   uint   `json:"id" binding:"required"`
	Team string `json:"team" binding:"required"`
}

// AssignEduRepairOrder 人工派单
// @Tags Eduadmin
// @Summary 人工派单（待派→处理中）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.assignRepairReq true "派单请求"
// @Success 200 {object} response.Response{data=model.EduRepairOrder} "派发成功"
// @Router /eduadmin/repair/assign [post]
func (a *EduadminApi) AssignEduRepairOrder(c *gin.Context) {
	var in assignRepairReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.AssignEduRepairOrder(in.ID, in.Team)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "派发成功", c)
}

// CloseEduRepairOrder 闭环
// @Tags Eduadmin
// @Summary 工单闭环（处理中→已闭环）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "工单ID"
// @Success 200 {object} response.Response{data=model.EduRepairOrder} "闭环成功"
// @Router /eduadmin/repair/close [post]
func (a *EduadminApi) CloseEduRepairOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	out, err := service.EduadminApp.CloseEduRepairOrder(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "闭环成功", c)
}
