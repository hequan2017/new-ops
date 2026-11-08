// Package router artsci 路由
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/middleware"

	v1 "github.com/hequan2017/new-ops/server/plugin/artsci/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 artsci 路由（private 组，JWT+casbin 由 initialize.Router 显式挂载）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	base := Router.Group("artsci")
	{
		base.GET("ping", v1.Api.Ping)
	}
	tool := Router.Group("artsci/tool")
	{
		tool.GET("list", v1.Api.GetTools)
		tool.POST("", middleware.OperationRecord(), v1.Api.SaveTool)
		tool.POST("run", middleware.OperationRecord(), v1.Api.RunTool)
		tool.GET("runLogs", v1.Api.GetRunLogs)
		tool.DELETE("", middleware.OperationRecord(), v1.Api.DeleteTool)
	}
	grading := Router.Group("artsci/grading")
	{
		grading.POST("", middleware.OperationRecord(), v1.Api.GradeEssay)
		grading.GET("list", v1.Api.GetGradings)
	}
	question := Router.Group("artsci/question")
	{
		question.POST("generate", middleware.OperationRecord(), v1.Api.GenerateQuestion)
		question.GET("list", v1.Api.GetQuestions)
	}
	paper := Router.Group("artsci/paper")
	{
		paper.POST("", middleware.OperationRecord(), v1.Api.BuildPaper)
		paper.GET("list", v1.Api.GetPapers)
	}
	scene := Router.Group("artsci/scene")
	{
		scene.GET("list", v1.Api.GetSceneTemplates)
	}
	img := Router.Group("artsci/imageTask")
	{
		img.POST("", middleware.OperationRecord(), v1.Api.SubmitImageTask)
		img.GET("list", v1.Api.GetImageTasks)
	}
	imgAsset := Router.Group("artsci/imageAsset")
	{
		imgAsset.POST("", middleware.OperationRecord(), v1.Api.RegisterImageAsset)
		imgAsset.GET("list", v1.Api.GetImageAssets)
		imgAsset.POST("interpret", middleware.OperationRecord(), v1.Api.InterpretImage)
	}
	interp := Router.Group("artsci/interpretation")
	{
		interp.GET("list", v1.Api.GetInterpretations)
	}
}
