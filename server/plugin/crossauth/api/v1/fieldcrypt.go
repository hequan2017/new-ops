// Package v1 crossauth API：敏感数据加密应用（需求 3.7-10）
package v1

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// fieldCryptReq 字段加解密请求
type fieldCryptReq struct {
	Algorithm string `json:"algorithm" binding:"required"`
	Plaintext string `json:"plaintext"`
	HexText   string `json:"hexText"`
}

// EncryptCaField 字段加密
// @Tags Crossauth
// @Summary 字段加密适配器（SM4/AES-GCM 可选，密文不回显明文）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.fieldCryptReq true "加密请求"
// @Success 200 {object} response.Response{data=map[string]any} "加密成功"
// @Router /crossauth/fieldCrypt/encrypt [post]
func (a *CrossauthApi) EncryptCaField(c *gin.Context) {
	var in fieldCryptReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.EncryptCaField(in.Algorithm, []byte(in.Plaintext))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"cipher": out, "algorithm": in.Algorithm}, "加密成功", c)
}

// DecryptCaField 字段解密
// @Tags Crossauth
// @Summary 字段解密适配器（SM4 hex / AES-GCM 密文串）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.fieldCryptReq true "解密请求"
// @Success 200 {object} response.Response{data=map[string]any} "解密成功"
// @Router /crossauth/fieldCrypt/decrypt [post]
func (a *CrossauthApi) DecryptCaField(c *gin.Context) {
	var in fieldCryptReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.CaApp.DecryptCaField(in.Algorithm, in.HexText)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"plaintext": string(out)}, "解密成功", c)
}

// watermarkReq 水印请求
type watermarkReq struct {
	Text     string `json:"text" binding:"required"`
	UserName string `json:"userName" binding:"required"`
}

// WatermarkCaData 访问水印
// @Tags Crossauth
// @Summary 敏感数据访问水印（用户+CST 时间戳嵌入返回文本）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.watermarkReq true "水印请求"
// @Success 200 {object} response.Response{data=map[string]any} "水印完成"
// @Router /crossauth/fieldCrypt/watermark [post]
func (a *CrossauthApi) WatermarkCaData(c *gin.Context) {
	var in watermarkReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out := service.WatermarkText(in.Text, in.UserName, time.Now())
	response.OkWithDetailed(gin.H{"watermarked": out}, "水印完成", c)
}
