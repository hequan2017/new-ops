// Package v1 crossauth API：MFA 管理（需求 3.7-3）
package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// enrollMfaReq MFA 登记请求
type enrollMfaReq struct {
	UserName string `json:"userName" binding:"required"`
	Method   string `json:"method" binding:"required"`
	RoleName string `json:"roleName"`
}

// EnrollCaMfa 登记 MFA
// @Tags Crossauth
// @Summary 登记 MFA（APP动态码签发 TOTP 密钥密文落库，返回一次明文供扫码）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.enrollMfaReq true "登记请求"
// @Success 200 {object} response.Response{data=map[string]any} "登记成功"
// @Router /crossauth/mfa [post]
func (a *CrossauthApi) EnrollCaMfa(c *gin.Context) {
	var in enrollMfaReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	cfg, secret, err := service.CaApp.EnrollCaMfa(in.UserName, in.Method, in.RoleName)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	out := gin.H{"config": cfg}
	if secret != "" {
		out["totpSecretOnce"] = secret // 仅签发时返回一次
	}
	response.OkWithDetailed(out, "登记成功", c)
}

// GetCaMfaConfigs 配置列表
// @Tags Crossauth
// @Summary MFA 配置列表（密钥不回显）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.CaMfaConfig} "获取成功"
// @Router /crossauth/mfa/list [get]
func (a *CrossauthApi) GetCaMfaConfigs(c *gin.Context) {
	list, err := service.CaApp.GetCaMfaConfigs()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// verifyMfaReq 校验请求
type verifyMfaReq struct {
	UserName string `json:"userName" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

// VerifyCaMfa 动态码校验
// @Tags Crossauth
// @Summary MFA 校验（APP动态码 TOTP ±1 窗口；其余方式口径留痕）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.verifyMfaReq true "校验请求"
// @Success 200 {object} response.Response{data=map[string]any} "校验完成"
// @Router /crossauth/mfa/verify [post]
func (a *CrossauthApi) VerifyCaMfa(c *gin.Context) {
	var in verifyMfaReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	ok, err := service.CaApp.VerifyCaMfa(in.UserName, in.Code)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"passed": ok}, "校验完成", c)
}

// SaveCaMfaPolicy 保存策略
// @Tags Crossauth
// @Summary 保存高权限角色强制 MFA 策略（按角色幂等）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.CaMfaPolicy true "策略"
// @Success 200 {object} response.Response{data=model.CaMfaPolicy} "保存成功"
// @Router /crossauth/mfa/policy [post]
func (a *CrossauthApi) SaveCaMfaPolicy(c *gin.Context) {
	var in model.CaMfaPolicy
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.SaveCaMfaPolicy(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetCaMfaPolicies 策略列表
// @Tags Crossauth
// @Summary 强制 MFA 策略列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.CaMfaPolicy} "获取成功"
// @Router /crossauth/mfa/policy/list [get]
func (a *CrossauthApi) GetCaMfaPolicies(c *gin.Context) {
	list, err := service.CaApp.GetCaMfaPolicies()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
