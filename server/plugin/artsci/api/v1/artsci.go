// Package v1 artsci API：文理学科 AI 工具集接口层（参数绑定与响应，禁止业务逻辑）
package v1

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
)

var Api = new(ArtsciApi)

type ArtsciApi struct{}

// Ping 插件连通性探测
// @Tags Artsci
// @Summary artsci 连通性探测
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]any} "探测成功"
// @Router /artsci/ping [get]
func (a *ArtsciApi) Ping(c *gin.Context) {
	response.OkWithDetailed(gin.H{
		"plugin": "artsci",
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}, "探测成功", c)
}
