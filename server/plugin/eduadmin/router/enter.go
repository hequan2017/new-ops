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
	asset := Router.Group("eduadmin/asset")
	{
		asset.GET("list", v1.Api.GetEduAssets)
		asset.GET("maintenance", v1.Api.PredictAssetFault)
		asset.GET("maintenance/list", v1.Api.GetEduMaintPlans)
		asset.POST("", middleware.OperationRecord(), v1.Api.SaveEduAsset)
		asset.POST("reading", middleware.OperationRecord(), v1.Api.AddEduSensorReading)
	}
	pur := Router.Group("eduadmin/purchase")
	{
		pur.GET("list", v1.Api.GetEduPurchases)
		pur.GET("suggest", v1.Api.SuggestEduPurchase)
		pur.POST("", middleware.OperationRecord(), v1.Api.SaveEduPurchase)
	}
	energy := Router.Group("eduadmin/energy")
	{
		energy.GET("list", v1.Api.GetEduEnergyRecords)
		energy.GET("analyze", v1.Api.AnalyzeEduEnergy)
		energy.POST("", middleware.OperationRecord(), v1.Api.SaveEduEnergyRecord)
	}
	meal := Router.Group("eduadmin/meal")
	{
		meal.GET("list", v1.Api.GetEduMealRecords)
		meal.GET("forecast", v1.Api.ForecastEduMeals)
		meal.POST("", middleware.OperationRecord(), v1.Api.SaveEduMealRecord)
	}
	repair := Router.Group("eduadmin/repair")
	{
		repair.GET("list", v1.Api.GetEduRepairOrders)
		repair.POST("", middleware.OperationRecord(), v1.Api.CreateEduRepairOrder)
		repair.POST("assign", middleware.OperationRecord(), v1.Api.AssignEduRepairOrder)
		repair.POST("close", middleware.OperationRecord(), v1.Api.CloseEduRepairOrder)
	}
	env := Router.Group("eduadmin/env")
	{
		env.GET("list", v1.Api.GetEduEnvReadings)
		env.GET("alerts", v1.Api.GetEduEnvAlerts)
		env.POST("", v1.Api.SaveEduEnvReading)
	}
	Router.Group("eduadmin/dashboard").GET("", v1.Api.GetEduDashboard)
	run := Router.Group("eduadmin/run")
	{
		run.GET("index/list", v1.Api.GetEduRunIndexes)
		run.GET("alerts", v1.Api.GetEduRunAlerts)
		run.POST("index", middleware.OperationRecord(), v1.Api.SaveEduRunIndex)
		run.POST("scan", middleware.OperationRecord(), v1.Api.ScanEduRunWarning)
	}
	policy := Router.Group("eduadmin/policy")
	{
		policy.POST("simulate", v1.Api.SimulateEduPolicy)
		policy.POST("compare", v1.Api.CompareEduPolicySims)
	}
	share := Router.Group("eduadmin/share")
	{
		share.GET("resource/list", v1.Api.GetEduShareResources)
		share.GET("practices", v1.Api.GetEduBestPractices)
		share.POST("resource", middleware.OperationRecord(), v1.Api.SaveEduShareResource)
		share.POST("schedule", middleware.OperationRecord(), v1.Api.ScheduleEduShareResource)
	}
	Router.Group("eduadmin/jointBudget").GET("", v1.Api.AnalyzeEduJointBudget)
	sync := Router.Group("eduadmin/sync")
	{
		sync.GET("adapter/list", v1.Api.GetEduSyncAdapters)
		sync.GET("task/list", v1.Api.GetEduSyncTasks)
		sync.POST("adapter", middleware.OperationRecord(), v1.Api.SaveEduSyncAdapter)
		sync.POST("run", middleware.OperationRecord(), v1.Api.RunEduSyncTask)
	}
	idp := Router.Group("eduadmin/idp")
	{
		idp.GET("list", v1.Api.GetEduIdpAdapters)
		idp.POST("", middleware.OperationRecord(), v1.Api.SaveEduIdpAdapter)
		idp.POST("redirect", v1.Api.BuildEduIdpRedirect)
		idp.POST("userSync", middleware.OperationRecord(), v1.Api.EduIdpUserSync)
	}
	gov := Router.Group("eduadmin/govPolicy")
	{
		gov.GET("list", v1.Api.GetEduGovPolicies)
		gov.POST("pull", middleware.OperationRecord(), v1.Api.PullEduGovPolicy)
	}
}
