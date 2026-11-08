// Package v1 fanhua API：繁花-文科大模型接口层（参数绑定与响应，禁止业务逻辑）
package v1

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
)

var Api = new(FanhuaApi)

type FanhuaApi struct{}

// Ping 插件连通性探测
// @Tags Fanhua
// @Summary fanhua 连通性探测
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]any} "探测成功"
// @Router /fanhua/ping [get]
func (a *FanhuaApi) Ping(c *gin.Context) {
	response.OkWithDetailed(gin.H{
		"plugin": "fanhua",
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}, "探测成功", c)
}
