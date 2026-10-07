// Package api 容器日志流与 exec 终端 WebSocket 端点
// 握手鉴权与 term 插件一致：浏览器 WS 无法带 header，token 经 query 传入校验。
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/hequan2017/new-ops/server/plugin/container/service"
	"github.com/hequan2017/new-ops/server/utils"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	// 内网工具，允许任意来源；生产应收敛到部署域
	CheckOrigin: func(r *http.Request) bool { return true },
}

// parseEndpointContainer 解析公共 query 参数
func parseEndpointContainer(c *gin.Context) (uint, string, bool) {
	token := c.Query("token")
	if token == "" {
		c.String(http.StatusUnauthorized, "缺少 token")
		return 0, "", false
	}
	if _, err := utils.NewJWT().ParseToken(token); err != nil {
		c.String(http.StatusUnauthorized, "token 无效: "+err.Error())
		return 0, "", false
	}
	endpointID, err := strconv.ParseUint(c.Query("endpointId"), 10, 64)
	if err != nil || endpointID == 0 {
		c.String(http.StatusBadRequest, "endpointId 无效")
		return 0, "", false
	}
	cid := c.Query("id")
	if cid == "" {
		c.String(http.StatusBadRequest, "容器 ID 不能为空")
		return 0, "", false
	}
	return uint(endpointID), cid, true
}

// LogsWS 容器日志流
// @Tags DockerContainer
// @Summary 容器日志流（WebSocket，docker logs -f）
// @Security ApiKeyAuth
// @Param token query string true "JWT"
// @Param endpointId query int true "接入点ID"
// @Param id query string true "容器ID"
// @Param tail query int false "初始回看行数（默认200，上限5000）"
// @Success 200 {string} string "升级为 WebSocket"
// @Router /container/container/logws [get]
func (a *containerApi) LogsWS(c *gin.Context) {
	endpointID, cid, ok := parseEndpointContainer(c)
	if !ok {
		return
	}
	tail, _ := strconv.Atoi(c.DefaultQuery("tail", "200"))
	ws, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	// 阻塞桥接：结束后 WS 已被关闭
	service.Service.Endpoint.LogsWS(ws, endpointID, cid, tail)
}

// ExecWS 容器 exec 交互终端
// @Tags DockerContainer
// @Summary 容器 exec 终端（WebSocket，/bin/sh）
// @Security ApiKeyAuth
// @Param token query string true "JWT"
// @Param endpointId query int true "接入点ID"
// @Param id query string true "容器ID"
// @Param cols query int false "列数"
// @Param rows query int false "行数"
// @Success 200 {string} string "升级为 WebSocket"
// @Router /container/container/execws [get]
func (a *containerApi) ExecWS(c *gin.Context) {
	endpointID, cid, ok := parseEndpointContainer(c)
	if !ok {
		return
	}
	cols, _ := strconv.Atoi(c.DefaultQuery("cols", "120"))
	rows, _ := strconv.Atoi(c.DefaultQuery("rows", "30"))
	ws, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	service.Service.Endpoint.ExecWS(ws, endpointID, cid, cols, rows, "")
}
