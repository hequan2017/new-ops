// Package api 凭据保险库接口（明文永不回显）
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/request"
	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
)

type credCredential struct{}

// CreateCredential 创建凭据
// @Tags Credential
// @Summary 创建凭据（AES-256-GCM 加密落库）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "name/type/username/secret/remark"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /asset/credential/create [post]
func (a *credCredential) CreateCredential(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Type     string `json:"type" binding:"required"`
		Username string `json:"username"`
		Secret   string `json:"secret" binding:"required"`
		Remark   string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	cred := &model.CredCredential{
		Name: req.Name, Type: req.Type, Username: req.Username, Remark: req.Remark,
	}
	if err := credCredentialService.CreateCredential(cred, req.Secret); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteCredential 删除凭据（被引用时拒绝）
// @Tags Credential
// @Summary 删除凭据
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.GetById true "凭据ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /asset/credential/delete [delete]
func (a *credCredential) DeleteCredential(c *gin.Context) {
	var req request.GetById
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := credCredentialService.DeleteCredential(uint(req.ID)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// UpdateCredential 更新凭据（secret 留空则保留原密文）
// @Tags Credential
// @Summary 更新凭据
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "ID/name/username/secret(可空)/remark"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /asset/credential/update [put]
func (a *credCredential) UpdateCredential(c *gin.Context) {
	var req struct {
		ID       uint   `json:"ID" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Username string `json:"username"`
		Secret   string `json:"secret"`
		Remark   string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	cred := &model.CredCredential{
		Name: req.Name, Username: req.Username, Remark: req.Remark,
	}
	cred.ID = req.ID
	if err := credCredentialService.UpdateCredential(cred, req.Secret); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// GetCredentialList 凭据列表（密文与明文均不返回）
// @Tags Credential
// @Summary 凭据列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param keyword query string false "名称/用户名/备注关键字"
// @Success 200 {object} response.Response{data=[]model.CredCredential} "获取成功"
// @Router /asset/credential/list [get]
func (a *credCredential) GetCredentialList(c *gin.Context) {
	list, err := credCredentialService.GetCredentialList(c.Query("keyword"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}
