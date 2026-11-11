// Package v1 crossauth API：资源访问权限匹配与告警（需求 3.7-11）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// CheckCaAccess 访问校验
// @Tags Crossauth
// @Summary 资源访问校验（矩阵+互认双门，拒绝即阻断并告警推送 归属校管理员+安全团队）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body service.accessCheckReq true "访问请求"
// @Success 200 {object} response.Response{data=map[string]any} "校验完成"
// @Router /crossauth/access/check [post]
func (a *CrossauthApi) CheckCaAccess(c *gin.Context) {
	var in service.AccessCheckReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.CheckCaAccess(in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "校验完成", c)
}

// GetCaAccessLogs 访问日志
// @Tags Crossauth
// @Summary 访问日志列表（用户过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param userName query string false "用户"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /crossauth/access/logs [get]
func (a *CrossauthApi) GetCaAccessLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.CaApp.GetCaAccessLogs(c.Query("userName"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetCaAlerts 告警列表
// @Tags Crossauth
// @Summary 阻断告警列表（处置状态过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param handled query string false "yes/no"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /crossauth/access/alerts [get]
func (a *CrossauthApi) GetCaAlerts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.CaApp.GetCaAlerts(c.Query("handled"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// HandleCaAlert 处置告警
// @Tags Crossauth
// @Summary 告警处置标记
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "告警ID"
// @Success 200 {object} response.Response "处置成功"
// @Router /crossauth/access/handle [post]
func (a *CrossauthApi) HandleCaAlert(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if err := service.CaApp.HandleCaAlert(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("处置成功", c)
}
