// Package router fanhua 路由
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/middleware"

	v1 "github.com/hequan2017/new-ops/server/plugin/fanhua/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 fanhua 路由（private 组，JWT+casbin 由 initialize.Router 显式挂载）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	base := Router.Group("fanhua")
	{
		base.GET("ping", v1.Api.Ping)
	}
	corpus := Router.Group("fanhua/corpus")
	{
		corpus.GET("list", v1.Api.GetCorpora)
		corpus.POST("", middleware.OperationRecord(), v1.Api.SaveCorpus)
		corpus.POST("import", middleware.OperationRecord(), v1.Api.ImportCorpora)
		corpus.DELETE("", middleware.OperationRecord(), v1.Api.DeleteCorpus)
	}
	para := Router.Group("fanhua/parallel")
	{
		para.GET("list", v1.Api.GetParallels)
		para.GET("search", v1.Api.SearchParallel)
		para.GET("export", v1.Api.ExportParallels)
		para.POST("", middleware.OperationRecord(), v1.Api.SaveParallel)
		para.DELETE("", middleware.OperationRecord(), v1.Api.DeleteParallel)
	}
	clean := Router.Group("fanhua/clean")
	{
		clean.POST("dry", v1.Api.CleanDry)
		clean.POST("", middleware.OperationRecord(), v1.Api.CleanCorpora)
	}
	Router.Group("fanhua").POST("tokenize", v1.Api.Tokenize)
	entity := Router.Group("fanhua/entity")
	{
		entity.GET("list", v1.Api.GetEntities)
		entity.GET("hits", v1.Api.GetEntityHits)
		entity.POST("annotate", middleware.OperationRecord(), v1.Api.AnnotateText)
		entity.POST("", middleware.OperationRecord(), v1.Api.SaveEntity)
		entity.DELETE("", middleware.OperationRecord(), v1.Api.DeleteEntity)
	}
	label := Router.Group("fanhua/label")
	{
		label.GET("list", v1.Api.GetTextLabels)
		label.GET("stats", v1.Api.LabelStats)
		label.POST("build", middleware.OperationRecord(), v1.Api.BuildTextLabels)
		label.PUT("", middleware.OperationRecord(), v1.Api.UpdateLabel)
	}
	imgAsset := Router.Group("fanhua/imageAsset")
	{
		imgAsset.GET("list", v1.Api.GetImageAssets)
		imgAsset.POST("", middleware.OperationRecord(), v1.Api.SaveImageAsset)
		imgAsset.POST("label", middleware.OperationRecord(), v1.Api.LabelImage)
	}
	Router.Group("fanhua/imageLabel").GET("list", v1.Api.GetImageLabels)
	voice := Router.Group("fanhua/voice")
	{
		voice.GET("list", v1.Api.GetVoiceEntries)
		voice.POST("search", v1.Api.SearchCulture)
		voice.POST("", middleware.OperationRecord(), v1.Api.SaveVoiceEntry)
		voice.DELETE("", middleware.OperationRecord(), v1.Api.DeleteVoiceEntry)
	}
	word := Router.Group("fanhua/word")
	{
		word.GET("list", v1.Api.GetWords)
		word.POST("", middleware.OperationRecord(), v1.Api.SaveWord)
		word.DELETE("", middleware.OperationRecord(), v1.Api.DeleteWord)
	}
	Router.Group("fanhua/values").POST("health", v1.Api.GetHealth)
	Router.Group("fanhua/template").GET("list", v1.Api.GetTemplates)
	Router.Group("fanhua/conflict").GET("list", v1.Api.GetConflictCases)
	gov := Router.Group("fanhua/govIndicator")
	{
		gov.GET("list", v1.Api.GetIndicatorTree)
		gov.POST("", middleware.OperationRecord(), v1.Api.SaveIndicator)
		gov.DELETE("", middleware.OperationRecord(), v1.Api.DeleteIndicator)
	}
	Router.Group("fanhua/govData").GET("list", v1.Api.GetDataPoints)
	Router.Group("fanhua/govData").POST("", middleware.OperationRecord(), v1.Api.SaveDataPoint)
	Router.Group("fanhua/govProfile").GET("", v1.Api.GetRegionProfile)
	Router.Group("fanhua/govPhoto").POST("", middleware.OperationRecord(), v1.Api.SaveGovPhoto)
	Router.Group("fanhua/govFusion").GET("", v1.Api.FusionQuery)
	prompt := Router.Group("fanhua/prompt")
	{
		prompt.GET("list", v1.Api.GetPromptTemplates)
		prompt.POST("render", v1.Api.RenderPrompt)
		prompt.POST("", middleware.OperationRecord(), v1.Api.SavePromptTemplate)
		prompt.DELETE("", middleware.OperationRecord(), v1.Api.DeletePromptTemplate)
	}
	qa := Router.Group("fanhua/dialectQa")
	{
		qa.POST("", middleware.OperationRecord(), v1.Api.DialectQA)
		qa.GET("logs", v1.Api.GetDialectQaLogs)
	}
	audit := Router.Group("fanhua/valuesAudit")
	{
		audit.POST("", middleware.OperationRecord(), v1.Api.AuditValues)
		audit.GET("logs", v1.Api.GetValuesAuditLogs)
	}
}
