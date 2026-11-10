// Package router eduadmin 路由
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/middleware"

	v1 "github.com/hequan2017/new-ops/server/plugin/eduadmin/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 eduadmin 路由（private 组，JWT+casbin 由 initialize.Router 显式挂载）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	base := Router.Group("eduadmin")
	{
		base.GET("ping", v1.Api.Ping)
	}
	rec := Router.Group("eduadmin/record")
	{
		rec.GET("list", v1.Api.GetEduRecords)
		rec.POST("import", middleware.OperationRecord(), v1.Api.ImportEduRecords)
		rec.POST("", middleware.OperationRecord(), v1.Api.SaveEduRecord)
		rec.DELETE("", middleware.OperationRecord(), v1.Api.DeleteEduRecord)
	}
	Router.Group("eduadmin/clean").POST("", middleware.OperationRecord(), v1.Api.CleanEduRecords)
	Router.Group("eduadmin/mask").POST("", v1.Api.MaskPayload)
	Router.Group("eduadmin/mask").GET("batch", v1.Api.MaskBatch)
	pol := Router.Group("eduadmin/policy")
	{
		pol.GET("list", v1.Api.GetEduPolicies)
		pol.GET("search", v1.Api.SearchEduPolicies)
		pol.POST("", middleware.OperationRecord(), v1.Api.SaveEduPolicy)
		pol.DELETE("", middleware.OperationRecord(), v1.Api.DeleteEduPolicy)
	}
	res := Router.Group("eduadmin/resume")
	{
		res.GET("list", v1.Api.GetEduResumes)
		res.POST("", middleware.OperationRecord(), v1.Api.SaveEduResume)
	}
	pos := Router.Group("eduadmin/position")
	{
		pos.GET("list", v1.Api.GetEduPositions)
		pos.POST("", middleware.OperationRecord(), v1.Api.SaveEduPosition)
	}
	Router.Group("eduadmin/screen").GET("", v1.Api.ScreenCandidates)
	perf := Router.Group("eduadmin/perf")
	{
		perf.GET("list", v1.Api.GetEduPerfRecords)
		perf.POST("", middleware.OperationRecord(), v1.Api.SaveEduPerfRecord)
		perf.POST("report", middleware.OperationRecord(), v1.Api.GeneratePerfReport)
		perf.GET("report/list", v1.Api.GetEduPerfReports)
	}
	talent := Router.Group("eduadmin/talent")
	{
		talent.GET("stat/list", v1.Api.GetEduTalentStats)
		talent.GET("forecast", v1.Api.ForecastTalent)
		talent.POST("stat", middleware.OperationRecord(), v1.Api.SaveEduTalentStat)
	}
	budget := Router.Group("eduadmin/budget")
	{
		budget.GET("list", v1.Api.GetEduBudgets)
		budget.GET("analysis", v1.Api.AnalyzeEduBudget)
		budget.GET("suggest", v1.Api.SuggestEduBudget)
		budget.POST("", middleware.OperationRecord(), v1.Api.SaveEduBudget)
		budget.POST("compare", v1.Api.CompareEduBudgetPlans)
	}
	exp := Router.Group("eduadmin/expense")
	{
		exp.GET("list", v1.Api.GetEduInvoices)
		exp.POST("", middleware.OperationRecord(), v1.Api.SubmitEduExpense)
		exp.POST("pay", middleware.OperationRecord(), v1.Api.PayEduExpense)
		exp.POST("reject", middleware.OperationRecord(), v1.Api.RejectEduExpense)
	}
	risk := Router.Group("eduadmin/risk")
	{
		risk.GET("list", v1.Api.GetEduRiskAlerts)
		risk.POST("scan", middleware.OperationRecord(), v1.Api.ScanEduRisk)
		risk.POST("handle", middleware.OperationRecord(), v1.Api.HandleEduRiskAlert)
	}
}
