// Package v1 crossauth API：四级权限矩阵（需求 3.7-7）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// SaveCaPermGrant 保存授权
// @Tags Crossauth
// @Summary 新增四级权限授权（校-部门-用户-场景，* 通配，空维度归一通配）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.CaPermGrant true "授权"
// @Success 200 {object} response.Response{data=model.CaPermGrant} "保存成功"
// @Router /crossauth/matrix [post]
func (a *CrossauthApi) SaveCaPermGrant(c *gin.Context) {
	var in model.CaPermGrant
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.SaveCaPermGrant(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeleteCaPermGrant 删除授权
// @Tags Crossauth
// @Summary 删除四级授权
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "授权ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /crossauth/matrix [delete]
func (a *CrossauthApi) DeleteCaPermGrant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.CaApp.DeleteCaPermGrant(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetCaPermGrants 授权列表
// @Tags Crossauth
// @Summary 四级授权列表（院校过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param school query string false "院校"
// @Success 200 {object} response.Response{data=[]model.CaPermGrant} "获取成功"
// @Router /crossauth/matrix/list [get]
func (a *CrossauthApi) GetCaPermGrants(c *gin.Context) {
	list, err := service.CaApp.GetCaPermGrants(c.Query("school"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// matrixCheckReq 矩阵校验请求
type matrixCheckReq struct {
	School string `json:"school"`
	Dept   string `json:"dept"`
	User   string `json:"user"`
	Scene  string `json:"scene" binding:"required"`
}

// CheckCaMatrix 矩阵校验
// @Tags Crossauth
// @Summary 四级权限校验（最具体优先，无命中默认拒绝）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.matrixCheckReq true "校验请求"
// @Success 200 {object} response.Response{data=map[string]any} "校验完成"
// @Router /crossauth/matrix/check [post]
func (a *CrossauthApi) CheckCaMatrix(c *gin.Context) {
	var in matrixCheckReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.CheckCaMatrix(in.School, in.Dept, in.User, in.Scene)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "校验完成", c)
}
