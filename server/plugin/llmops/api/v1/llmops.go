// Package v1 llmops API：大模型应用开发平台接口层（参数绑定与响应，禁止业务逻辑）
package v1

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
)

var Api = new(LlmopsApi)

type LlmopsApi struct{}

// Ping 插件连通性探测
// @Tags Llmops
// @Summary llmops 连通性探测
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]any} "探测成功"
// @Router /llmops/ping [get]
func (a *LlmopsApi) Ping(c *gin.Context) {
	response.OkWithDetailed(gin.H{
		"plugin": "llmops",
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}, "探测成功", c)
}
