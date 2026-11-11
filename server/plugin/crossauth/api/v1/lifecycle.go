// Package v1 crossauth API：权限生命周期（需求 3.7-8）
package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// SaveCaGrantExpiry 登记有效期
// @Tags Crossauth
// @Summary 登记授权有效期（用户×角色幂等，yyyy-MM-dd 到期日）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.CaGrantExpiry true "有效期"
// @Success 200 {object} response.Response{data=model.CaGrantExpiry} "保存成功"
// @Router /crossauth/lifecycle [post]
func (a *CrossauthApi) SaveCaGrantExpiry(c *gin.Context) {
	var in model.CaGrantExpiry
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.SaveCaGrantExpiry(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// ScanCaGrantExpiry 到期扫描
// @Tags Crossauth
// @summary 到期扫描（回收+到期前 7 天提醒，幂等）
// @Tags Crossauth
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]any} "扫描完成"
// @Router /crossauth/lifecycle/scan [post]
func (a *CrossauthApi) ScanCaGrantExpiry(c *gin.Context) {
	reminders, revoked, err := service.CaApp.ScanCaGrantExpiry()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"reminders": reminders, "revoked": revoked}, "扫描完成", c)
}

// GetCaGrantExpiries 有效期列表
// @Tags Crossauth
// @Summary 授权有效期列表（用户过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param userName query string false "用户"
// @Success 200 {object} response.Response{data=[]model.CaGrantExpiry} "获取成功"
// @Router /crossauth/lifecycle/list [get]
func (a *CrossauthApi) GetCaGrantExpiries(c *gin.Context) {
	list, err := service.CaApp.GetCaGrantExpiries(c.Query("userName"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// GetCaExpiryReminders 提醒列表
// @Tags Crossauth
// @Summary 到期提醒列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.CaExpiryReminder} "获取成功"
// @Router /crossauth/lifecycle/reminders [get]
func (a *CrossauthApi) GetCaExpiryReminders(c *gin.Context) {
	list, err := service.CaApp.GetCaExpiryReminders()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
