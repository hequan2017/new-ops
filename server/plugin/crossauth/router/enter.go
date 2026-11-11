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
	grant := Router.Group("crossauth/grant")
	{
		grant.GET("requests", v1.Api.GetCaGrantRequests)
		grant.GET("assignments", v1.Api.GetCaGrantAssignments)
		grant.POST("rule", middleware.OperationRecord(), v1.Api.SaveCaAutoGrantRule)
		grant.POST("auto", v1.Api.AutoGrantCa)
		grant.POST("request", v1.Api.SubmitCaGrantRequest)
		grant.POST("approve", middleware.OperationRecord(), v1.Api.ApproveCaGrantRequest)
		grant.POST("reject", middleware.OperationRecord(), v1.Api.RejectCaGrantRequest)
	}
	mtx := Router.Group("crossauth/matrix")
	{
		mtx.GET("list", v1.Api.GetCaPermGrants)
		mtx.POST("check", v1.Api.CheckCaMatrix)
		mtx.POST("", middleware.OperationRecord(), v1.Api.SaveCaPermGrant)
		mtx.DELETE("", middleware.OperationRecord(), v1.Api.DeleteCaPermGrant)
	}
	lc := Router.Group("crossauth/lifecycle")
	{
		lc.GET("list", v1.Api.GetCaGrantExpiries)
		lc.GET("reminders", v1.Api.GetCaExpiryReminders)
		lc.POST("", middleware.OperationRecord(), v1.Api.SaveCaGrantExpiry)
		lc.POST("scan", middleware.OperationRecord(), v1.Api.ScanCaGrantExpiry)
	}
	cls := Router.Group("crossauth/classify")
	{
		cls.GET("list", v1.Api.GetCaClassifyResults)
		cls.POST("scan", v1.Api.ScanCaClassify)
	}
	sm4 := Router.Group("crossauth/sm4")
	{
		sm4.GET("keys", v1.Api.GetCaKeyRotations)
		sm4.POST("rotate", middleware.OperationRecord(), v1.Api.RotateCaKey)
		sm4.POST("encrypt", v1.Api.EncryptCaData)
		sm4.POST("decrypt", v1.Api.DecryptCaData)
	}
	fc := Router.Group("crossauth/fieldCrypt")
	{
		fc.POST("encrypt", v1.Api.EncryptCaField)
		fc.POST("decrypt", v1.Api.DecryptCaField)
		fc.POST("watermark", v1.Api.WatermarkCaData)
	}
	acc := Router.Group("crossauth/access")
	{
		acc.POST("check", v1.Api.CheckCaAccess)
		acc.GET("logs", v1.Api.GetCaAccessLogs)
		acc.GET("alerts", v1.Api.GetCaAlerts)
		acc.POST("handle", middleware.OperationRecord(), v1.Api.HandleCaAlert)
	}
	comp := Router.Group("crossauth/compute")
	{
		comp.POST("usage", v1.Api.RecordCaComputeUsage)
		comp.GET("usage/list", v1.Api.GetCaComputeUsages)
		comp.GET("alerts", v1.Api.GetCaComputeAlerts)
		comp.POST("handle", middleware.OperationRecord(), v1.Api.HandleCaComputeAlert)
	}
	Router.Group("crossauth/monitor").GET("fiveDim", v1.Api.GetCaFiveDim)
	ae := Router.Group("crossauth/auditEvent")
	{
		ae.GET("list", v1.Api.GetCaAuditEvents)
		ae.POST("", v1.Api.RecordCaAuditEvent)
	}
	ops := Router.Group("crossauth/opsTicket")
	{
		ops.GET("list", v1.Api.GetCaOpsTickets)
		ops.POST("", middleware.OperationRecord(), v1.Api.CreateCaOpsTicket)
		ops.POST("advance", middleware.OperationRecord(), v1.Api.AdvanceCaOpsTicket)
		ops.POST("message", v1.Api.PostCaOpsMessage)
		ops.GET("messages", v1.Api.GetCaOpsMessages)
	}
	ra := Router.Group("crossauth/remoteAssist")
	{
		ra.POST("", middleware.OperationRecord(), v1.Api.GrantCaRemoteAssist)
		ra.POST("revoke", middleware.OperationRecord(), v1.Api.RevokeCaRemoteAssist)
		ra.GET("list", v1.Api.GetCaRemoteAssists)
	}
}
