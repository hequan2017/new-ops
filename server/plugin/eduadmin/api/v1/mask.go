// Package v1 eduadmin API：数据脱敏与分级（需求 3.6-3）
package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// maskReq 脱敏请求
type maskReq struct {
	Payload string `json:"payload" binding:"required"`
	Domain  string `json:"domain"`
}

// MaskPayload 单条脱敏
// @Tags Eduadmin
// @Summary 数据脱敏与分级（身份证/银行卡 SHA-256 不可逆、金额小数隐藏、四级分级标记）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.maskReq true "脱敏请求"
// @Success 200 {object} response.Response{data=service.MaskResult} "脱敏完成"
// @Router /eduadmin/mask [post]
func (a *EduadminApi) MaskPayload(c *gin.Context) {
	var in maskReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.MaskPayload(in.Payload)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "脱敏完成", c)
}

// MaskBatch 域内批量脱敏预览
// @Tags Eduadmin
// @Summary 域内数据批量脱敏预览（dry 不落库）
// @Security ApiKeyAuth
// @Produce application/json
// @Param domain query string true "人事/财务/资产/后勤"
// @Success 200 {object} response.Response{data=[]map[string]any} "预览完成"
// @Router /eduadmin/mask/batch [get]
func (a *EduadminApi) MaskBatch(c *gin.Context) {
	out, err := service.EduadminApp.MaskEduRecordBatch(c.Query("domain"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "预览完成", c)
}
