// Package api 白泽终端接口：WebSSH WebSocket 端点
// 握手鉴权：浏览器 WebSocket 无法携带自定义 header，token 经 query 传入，
// 复用底座 utils.NewJWT().ParseToken 校验后再升级连接（不经过 Casbin，终端权限=登录可用）。
package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/hequan2017/new-ops/server/plugin/term/service"
	"github.com/hequan2017/new-ops/server/utils"
)

type terminal struct{}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	// 测试环境为内网工具，允许任意来源；生产应收敛到部署域
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WebSSH 终端 WebSocket 端点
// @Tags Term
// @Summary WebSSH 终端（WebSocket）
// @Security ApiKeyAuth
// @Param token query string true "JWT"
// @Param hostId query int true "主机ID"
// @Param credentialId query int true "SSH凭据ID"
// @Param cols query int false "终端列数"
// @Param rows query int false "终端行数"
// @Success 200 {string} string "升级为 WebSocket"
// @Router /term/ws [get]
func (t *terminal) WebSSH(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.String(http.StatusUnauthorized, "缺少 token")
		return
	}
	claims, err := utils.NewJWT().ParseToken(token)
	if err != nil {
		c.String(http.StatusUnauthorized, "token 无效: "+err.Error())
		return
	}
	hostID, err := strconv.ParseUint(c.Query("hostId"), 10, 64)
	if err != nil || hostID == 0 {
		c.String(http.StatusBadRequest, "hostId 无效")
		return
	}
	credID, err := strconv.ParseUint(c.Query("credentialId"), 10, 64)
	if err != nil || credID == 0 {
		c.String(http.StatusBadRequest, "credentialId 无效")
		return
	}
	cols, _ := strconv.Atoi(c.DefaultQuery("cols", "120"))
	rows, _ := strconv.Atoi(c.DefaultQuery("rows", "30"))

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return // upgrade 失败时 gin 已写响应
	}
	params := service.TermStartParams{
		HostID:       uint(hostID),
		CredentialID: uint(credID),
		Cols:         cols,
		Rows:         rows,
		Operator:     claims.Username,
	}
	// 阻塞桥接：结束后 WS 已被关闭
	_ = service.TermService.StartWebSSH(ws, params)
}

// LogTail 远程日志 tail WebSocket 端点
// @Tags Term
// @Summary 远程日志 tail（WebSocket，只读流）
// @Security ApiKeyAuth
// @Param token query string true "JWT"
// @Param hostId query int true "主机ID"
// @Param credentialId query int true "SSH凭据ID"
// @Param path query string true "日志绝对路径"
// @Param lines query int false "初始回看行数（默认200，上限2000）"
// @Success 200 {string} string "升级为 WebSocket"
// @Router /term/logtail [get]
func (t *terminal) LogTail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.String(http.StatusUnauthorized, "缺少 token")
		return
	}
	claims, err := utils.NewJWT().ParseToken(token)
	if err != nil {
		c.String(http.StatusUnauthorized, "token 无效: "+err.Error())
		return
	}
	hostID, err := strconv.ParseUint(c.Query("hostId"), 10, 64)
	if err != nil || hostID == 0 {
		c.String(http.StatusBadRequest, "hostId 无效")
		return
	}
	credID, err := strconv.ParseUint(c.Query("credentialId"), 10, 64)
	if err != nil || credID == 0 {
		c.String(http.StatusBadRequest, "credentialId 无效")
		return
	}
	pathParam := c.Query("path")
	if pathParam == "" || pathParam[0] != '/' {
		c.String(http.StatusBadRequest, "path 必须为绝对路径")
		return
	}
	lines, _ := strconv.Atoi(c.DefaultQuery("lines", "200"))

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	params := service.LogTailParams{
		HostID:       uint(hostID),
		CredentialID: uint(credID),
		Path:         pathParam,
		Lines:        lines,
		Operator:     claims.Username,
	}
	_ = service.TermService.StartLogTail(ws, params)
}

// 提示：handler 中 fmt 引用由路由组注册使用
var _ = fmt.Sprintf
