// Package router exam 路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/exam/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 exam 路由（private 组，JWT+casbin 由 initialize.Router 显式挂载）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	sc := Router.Group("exam/score")
	{
		sc.POST("publish", v1.Api.PublishScores)
		sc.GET("list", v1.Api.GetScores)
		sc.GET("export", v1.Api.ExportScores)
	}
	gr := Router.Group("exam/grade")
	{
		gr.POST("run", v1.Api.RunGrade)
		gr.GET("list", v1.Api.GetGrades)
	}
	aw := Router.Group("exam/answer")
	{
		aw.POST("login", v1.Api.LoginByTicket)
		aw.GET("questions", v1.Api.GetSheetQuestions)
		aw.POST("save", v1.Api.SaveAnswers)
		aw.POST("submit", v1.Api.SubmitSheet)
	}
	cd := Router.Group("exam/candidate")
	{
		cd.GET("list", v1.Api.GetCandidates)
		cd.POST("import", v1.Api.ImportCandidates)
		cd.POST("claim", v1.Api.ClaimTicket)
		cd.DELETE("", v1.Api.DeleteCandidate)
	}
	e := Router.Group("exam/exam")
	{
		e.GET("list", v1.Api.GetExams)
		e.POST("", v1.Api.SaveExam)
		e.POST("transition", v1.Api.TransitionExam)
		e.DELETE("", v1.Api.DeleteExam)
	}
	p := Router.Group("exam/paper")
	{
		p.GET("list", v1.Api.GetPapers)
		p.POST("", v1.Api.SavePaper)
		p.POST("compose", v1.Api.ComposePaper)
		p.DELETE("", v1.Api.DeletePaper)
	}
	q := Router.Group("exam/question")
	{
		q.GET("list", v1.Api.GetQuestions)
		q.POST("", v1.Api.SaveQuestion)
		q.POST("import", v1.Api.ImportQuestions)
		q.DELETE("", v1.Api.DeleteQuestion)
	}
}
