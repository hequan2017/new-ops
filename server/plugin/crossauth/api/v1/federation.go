// Package v1 crossauth API：跨校身份系统对接（需求 3.7-1）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// saveFederationReq 对接配置请求（密钥明文仅入口，密文落库不回显）
type saveFederationReq struct {
	model.CaFederation
	Secret string `json:"secret"`
}

// SaveCaFederation 保存对接配置
// @Tags Crossauth
// @Summary 新建/更新高校身份对接（SAML2.0/OAuth2.0，密钥 AES-GCM 密文落库）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.saveFederationReq true "对接配置"
// @Success 200 {object} response.Response{data=model.CaFederation} "保存成功"
// @Router /crossauth/federation [post]
func (a *CrossauthApi) SaveCaFederation(c *gin.Context) {
	var in saveFederationReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.SaveCaFederation(&in.CaFederation, in.Secret)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetCaFederations 对接列表
// @Tags Crossauth
// @Summary 高校身份对接列表（密钥不回显）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.CaFederation} "获取成功"
// @Router /crossauth/federation/list [get]
func (a *CrossauthApi) GetCaFederations(c *gin.Context) {
	list, err := service.CaApp.GetCaFederations()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// GetCaSyncUsers 同步用户列表
// @Tags Crossauth
// @Summary 同步用户列表（院校/身份状态过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param school query string false "院校"
// @Param identityStatus query string false "在读/在职/离校/离职"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /crossauth/syncUser/list [get]
func (a *CrossauthApi) GetCaSyncUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.CaApp.GetCaSyncUsers(c.Query("school"), c.Query("identityStatus"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// RunCaSyncTask 执行同步
// @Tags Crossauth
// @Summary 执行用户同步任务（HTTP 拉取→身份状态白名单校验→幂等 upsert→执行记录）
// @Security ApiKeyAuth
// @Produce application/json
// @Param federationId query int true "对接配置ID"
// @Success 200 {object} response.Response{data=model.CaSyncTask} "执行完成"
// @Router /crossauth/federation/sync [post]
func (a *CrossauthApi) RunCaSyncTask(c *gin.Context) {
	federationId, _ := strconv.ParseUint(c.Query("federationId"), 10, 64)
	out, err := service.CaApp.RunCaSyncTask(uint(federationId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "执行完成", c)
}

// GetCaSyncTasks 任务记录
// @Tags Crossauth
// @Summary 同步任务执行记录（对接配置过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param federationId query int false "对接配置ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /crossauth/federation/tasks [get]
func (a *CrossauthApi) GetCaSyncTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	federationId, _ := strconv.ParseUint(c.Query("federationId"), 10, 64)
	list, total, err := service.CaApp.GetCaSyncTasks(uint(federationId), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}
