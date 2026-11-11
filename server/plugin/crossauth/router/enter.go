// Package router crossauth 路由
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/middleware"

	v1 "github.com/hequan2017/new-ops/server/plugin/crossauth/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 crossauth 路由（private 组，JWT+casbin 由 initialize.Router 显式挂载）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	base := Router.Group("crossauth")
	{
		base.GET("ping", v1.Api.Ping)
	}
	fed := Router.Group("crossauth/federation")
	{
		fed.GET("list", v1.Api.GetCaFederations)
		fed.GET("tasks", v1.Api.GetCaSyncTasks)
		fed.POST("", middleware.OperationRecord(), v1.Api.SaveCaFederation)
		fed.POST("sync", middleware.OperationRecord(), v1.Api.RunCaSyncTask)
	}
	Router.Group("crossauth/syncUser").GET("list", v1.Api.GetCaSyncUsers)
	trust := Router.Group("crossauth/trust")
	{
		trust.GET("list", v1.Api.GetCaTrustRules)
		trust.POST("check", v1.Api.CheckCaTrust)
		trust.POST("", middleware.OperationRecord(), v1.Api.SaveCaTrustRule)
		trust.DELETE("", middleware.OperationRecord(), v1.Api.DeleteCaTrustRule)
	}
	mfa := Router.Group("crossauth/mfa")
	{
		mfa.GET("list", v1.Api.GetCaMfaConfigs)
		mfa.GET("policy/list", v1.Api.GetCaMfaPolicies)
		mfa.POST("", v1.Api.EnrollCaMfa)
		mfa.POST("verify", v1.Api.VerifyCaMfa)
		mfa.POST("policy", middleware.OperationRecord(), v1.Api.SaveCaMfaPolicy)
	}
	sso := Router.Group("crossauth/sso")
	{
		sso.GET("apps", v1.Api.GetCaSsoApps)
		sso.POST("ticket", v1.Api.IssueCaSsoTicket)
		sso.POST("redeem", v1.Api.RedeemCaSsoTicket)
		sso.POST("app", middleware.OperationRecord(), v1.Api.SaveCaSsoApp)
	}
	role := Router.Group("crossauth/roleTemplate")
	{
		role.GET("list", v1.Api.GetCaRoleTemplates)
		role.POST("", middleware.OperationRecord(), v1.Api.SaveCaRoleTemplate)
	}
}
