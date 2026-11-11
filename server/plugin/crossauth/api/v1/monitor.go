// Package v1 crossauth API：五维监控看板（需求 3.7-13）
package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/service"
)

// GetCaFiveDim 五维看板
// @Tags Crossauth
// @Summary 五维监控看板（身份认证/权限变更/数据传输/资源使用/安全事件）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=service.CaFiveDim} "获取成功"
// @Router /crossauth/monitor/fiveDim [get]
func (a *CrossauthApi) GetCaFiveDim(c *gin.Context) {
	out, err := service.CaApp.GetCaFiveDim()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "获取成功", c)
}
