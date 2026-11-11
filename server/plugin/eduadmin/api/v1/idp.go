// Package v1 eduadmin API：跨校身份认证接口（需求 3.6-34）
package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SaveEduIdpAdapter 保存适配器
// @Tags Eduadmin
// @Summary 新建/更新跨校身份适配器（SAML2.0/OAuth2.0 元数据+互认范围）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduIdpAdapter true "适配器"
// @Success 200 {object} response.Response{data=model.EduIdpAdapter} "保存成功"
// @Router /eduadmin/idp [post]
func (a *EduadminApi) SaveEduIdpAdapter(c *gin.Context) {
	var in model.EduIdpAdapter
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduIdpAdapter(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetEduIdpAdapters 适配器列表
// @Tags Eduadmin
// @Summary 共建高校身份适配器列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.EduIdpAdapter} "获取成功"
// @Router /eduadmin/idp/list [get]
func (a *EduadminApi) GetEduIdpAdapters(c *gin.Context) {
	list, err := service.EduadminApp.GetEduIdpAdapters()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// idpRedirectReq 跳转请求
type idpRedirectReq struct {
	School string `json:"school" binding:"required"`
	State  string `json:"state" binding:"required"`
}

// BuildEduIdpRedirect 认证跳转
// @Tags Eduadmin
// @Summary 认证跳转 URL 组装（SAML RelayState / OAuth2.0 authorize）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.idpRedirectReq true "跳转请求"
// @Success 200 {object} response.Response{data=map[string]any} "组装成功"
// @Router /eduadmin/idp/redirect [post]
func (a *EduadminApi) BuildEduIdpRedirect(c *gin.Context) {
	var in idpRedirectReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.BuildEduIdpRedirect(in.School, in.State)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"redirectUrl": out}, "组装成功", c)
}

// EduIdpUserSync 用户同步任务
// @Tags Eduadmin
// @Summary 用户同步任务记录（同步频率 ≤5 分钟口径留痕）
// @Security ApiKeyAuth
// @Produce application/json
// @Param school query string true "高校"
// @Success 200 {object} response.Response{data=model.EduIdpUserSync} "执行完成"
// @Router /eduadmin/idp/userSync [post]
func (a *EduadminApi) EduIdpUserSync(c *gin.Context) {
	out, err := service.EduadminApp.RunEduIdpUserSync(c.Query("school"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "执行完成", c)
}
