// Package v1 crossauth API：身份互认规则库（需求 3.7-2）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// SaveCaTrustRule 保存规则
// @Tags Crossauth
// @Summary 新建/更新互认规则（本校身份×目标资源）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.CaTrustRule true "规则"
// @Success 200 {object} response.Response{data=model.CaTrustRule} "保存成功"
// @Router /crossauth/trust [post]
func (a *CrossauthApi) SaveCaTrustRule(c *gin.Context) {
	var in model.CaTrustRule
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.SaveCaTrustRule(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeleteCaTrustRule 删除规则
// @Tags Crossauth
// @Summary 删除互认规则
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "规则ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /crossauth/trust [delete]
func (a *CrossauthApi) DeleteCaTrustRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.CaApp.DeleteCaTrustRule(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetCaTrustRules 规则列表
// @Tags Crossauth
// @Summary 互认规则列表（身份过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param identity query string false "本校身份"
// @Success 200 {object} response.Response{data=[]model.CaTrustRule} "获取成功"
// @Router /crossauth/trust/list [get]
func (a *CrossauthApi) GetCaTrustRules(c *gin.Context) {
	list, err := service.CaApp.GetCaTrustRules(c.Query("identity"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// trustCheckReq 校验请求
type trustCheckReq struct {
	Identity string `json:"identity" binding:"required"`
	Resource string `json:"resource" binding:"required"`
}

// CheckCaTrust 互认校验
// @Tags Crossauth
// @Summary 身份互认校验（规则匹配+allowed 判定，输出放行结论）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.trustCheckReq true "校验请求"
// @Success 200 {object} response.Response{data=map[string]any} "校验完成"
// @Router /crossauth/trust/check [post]
func (a *CrossauthApi) CheckCaTrust(c *gin.Context) {
	var in trustCheckReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.CheckCaTrust(in.Identity, in.Resource)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "校验完成", c)
}
