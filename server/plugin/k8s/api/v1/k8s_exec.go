// Package api K2 收官：Pod WebShell WebSocket 端点
// 握手鉴权与 term/container 插件一致：浏览器 WS 无法带 header，token 经 query 传入校验。
package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/hequan2017/new-ops/server/utils"
)

var k8sWsUpgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	// 内网工具，允许任意来源；生产应收敛到部署域
	CheckOrigin: func(r *http.Request) bool { return true },
}

// PodExecWS Pod 交互终端
// @Tags K8sResource
// @Summary Pod WebShell（WebSocket，/bin/sh TTY）
// @Security ApiKeyAuth
// @Param token query string true "JWT"
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param pod query string true "Pod 名"
// @Param container query string false "容器名（空=Pod 首个容器）"
// @Param cols query int false "列数（默认120）"
// @Param rows query int false "行数（默认30）"
// @Success 200 {string} string "升级为 WebSocket"
// @Router /k8s/pod/execws [get]
func (t *k8sResource) PodExecWS(c *gin.Context) {
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
	// 终端为写级能力：握手时强制 casbin（query token 场景中间件不生效，需自验角色）
	if ok, _ := utils.GetCasbin().Enforce(fmt.Sprintf("%d", claims.AuthorityId), "/k8s/pod/execws", "GET"); !ok {
		c.String(http.StatusForbidden, "无终端使用权限")
		return
	}
	clusterID, err := strconv.ParseUint(c.Query("clusterId"), 10, 64)
	if err != nil || clusterID == 0 {
		c.String(http.StatusBadRequest, "clusterId 无效")
		return
	}
	ns := c.Query("namespace")
	pod := c.Query("pod")
	if ns == "" || pod == "" {
		c.String(http.StatusBadRequest, "namespace/pod 必填")
		return
	}
	cols, _ := strconv.Atoi(c.DefaultQuery("cols", "120"))
	rows, _ := strconv.Atoi(c.DefaultQuery("rows", "30"))
	ws, err := k8sWsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	// 阻塞桥接：结束后 WS 已被关闭
	k8sClusterService.PodExecWS(ws, uint(clusterID), ns, pod, c.Query("container"), cols, rows)
}
