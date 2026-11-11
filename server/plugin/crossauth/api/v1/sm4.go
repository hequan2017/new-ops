// Package v1 crossauth API：SM4 加密模块与密钥轮换（需求 3.7-10）
package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// rotateKeyReq 轮换请求
type rotateKeyReq struct {
	KeyName   string `json:"keyName" binding:"required"`
	CycleDays int    `json:"cycleDays"`
}

// RotateCaKey 轮换密钥
// @Tags Crossauth
// @Summary 密钥轮换（生成新 16 字节数据密钥，90 天周期可配，密文落库不回显）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.rotateKeyReq true "轮换请求"
// @Success 200 {object} response.Response{data=model.CaKeyRotation} "轮换成功"
// @Router /crossauth/sm4/rotate [post]
func (a *CrossauthApi) RotateCaKey(c *gin.Context) {
	var in rotateKeyReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.SaveCaKeyRotation(in.KeyName, in.CycleDays)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "轮换成功", c)
}

// GetCaKeyRotations 轮换表
// @Tags Crossauth
// @Summary 密钥轮换表列表（密钥不回显）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.CaKeyRotation} "获取成功"
// @Router /crossauth/sm4/keys [get]
func (a *CrossauthApi) GetCaKeyRotations(c *gin.Context) {
	list, err := service.CaApp.GetCaKeyRotations()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// cryptReq 加解密请求
type cryptReq struct {
	KeyName   string `json:"keyName" binding:"required"`
	Plaintext string `json:"plaintext"`
	HexText   string `json:"hexText"`
}

// EncryptCaData SM4 加密
// @Tags Crossauth
// @Summary SM4-ECB 加密（PKCS#7 填充，返回 hex）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.cryptReq true "加密请求"
// @Success 200 {object} response.Response{data=map[string]any} "加密成功"
// @Router /crossauth/sm4/encrypt [post]
func (a *CrossauthApi) EncryptCaData(c *gin.Context) {
	var in cryptReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.EncryptCaData(in.KeyName, []byte(in.Plaintext))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"hex": out}, "加密成功", c)
}

// DecryptCaData SM4 解密
// @Tags Crossauth
// @Summary SM4-ECB 解密（hex 入参）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.cryptReq true "解密请求"
// @Success 200 {object} response.Response{data=map[string]any} "解密成功"
// @Router /crossauth/sm4/decrypt [post]
func (a *CrossauthApi) DecryptCaData(c *gin.Context) {
	var in cryptReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.DecryptCaData(in.KeyName, in.HexText)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"plaintext": string(out)}, "解密成功", c)
}
