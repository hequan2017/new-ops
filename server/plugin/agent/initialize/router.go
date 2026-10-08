// Package initialize 白泽 Agent 通道路由
package initialize

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/middleware"
	agentV1 "github.com/hequan2017/new-ops/server/plugin/agent/api/v1"
	agentSvc "github.com/hequan2017/new-ops/server/plugin/agent/service"
)

// Router 注册 Agent 路由：REST 走 private（显式挂 JWT+casbin）；WS 走裸注册（鉴权=首帧注册令牌）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	private.GET("agent/instance/list", agentV1.AgentApi.ListInstances)
	private.POST("agent/instance/token", agentV1.AgentApi.IssueToken)
	private.DELETE("agent/instance", agentV1.AgentApi.RevokeInstance)

	// Agent 为非 JWT 主体（单文件二进制，注册令牌鉴权）：直接挂 http handler，不经 JWT 中间件组
	engine.GET(global.GVA_CONFIG.System.RouterPrefix+"/agent/ws", gin.WrapF(agentSvc.AgentSvc.HandleWS))
}
