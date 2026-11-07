// Package router exam 路由
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/middleware"

	v1 "github.com/hequan2017/new-ops/server/plugin/exam/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 exam 路由（private 组，JWT+casbin 由 initialize.Router 显式挂载）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	ac := Router.Group("exam/anticheat")
	{
		ac.POST("report", middleware.OperationRecord(), v1.Api.ReportBehavior)
		ac.GET("stats", v1.Api.GetBehaviorStats)
		ac.POST("check", middleware.OperationRecord(), v1.Api.CheckSimilarity)
	}
	nd := Router.Group("exam/node")
	{
		nd.POST("register", middleware.OperationRecord(), v1.Api.RegisterNode)
		nd.POST("heartbeat", v1.Api.NodeHeartbeat)
		nd.GET("list", v1.Api.GetNodes)
		nd.POST("assign", middleware.OperationRecord(), v1.Api.AssignShard)
	}
	sc := Router.Group("exam/score")
	{
		sc.POST("publish", middleware.OperationRecord(), v1.Api.PublishScores)
		sc.GET("list", v1.Api.GetScores)
		sc.GET("export", v1.Api.ExportScores)
	}
	gr := Router.Group("exam/grade")
	{
		gr.POST("run", middleware.OperationRecord(), v1.Api.RunGrade)
		gr.GET("list", v1.Api.GetGrades)
	}
	aw := Router.Group("exam/answer")
	{
		aw.POST("login", v1.Api.LoginByTicket)
		aw.GET("questions", v1.Api.GetSheetQuestions)
		aw.POST("save", middleware.OperationRecord(), v1.Api.SaveAnswers)
		aw.POST("submit", middleware.OperationRecord(), v1.Api.SubmitSheet)
	}
	cd := Router.Group("exam/candidate")
	{
		cd.GET("list", v1.Api.GetCandidates)
		cd.POST("import", middleware.OperationRecord(), v1.Api.ImportCandidates)
		cd.POST("claim", v1.Api.ClaimTicket)
		cd.DELETE("", middleware.OperationRecord(), v1.Api.DeleteCandidate)
	}
	e := Router.Group("exam/exam")
	{
		e.GET("list", v1.Api.GetExams)
		e.POST("", middleware.OperationRecord(), v1.Api.SaveExam)
		e.POST("transition", middleware.OperationRecord(), v1.Api.TransitionExam)
		e.DELETE("", middleware.OperationRecord(), v1.Api.DeleteExam)
	}
	p := Router.Group("exam/paper")
	{
		p.GET("list", v1.Api.GetPapers)
		p.POST("", middleware.OperationRecord(), v1.Api.SavePaper)
		p.POST("compose", middleware.OperationRecord(), v1.Api.ComposePaper)
		p.DELETE("", middleware.OperationRecord(), v1.Api.DeletePaper)
	}
	q := Router.Group("exam/question")
	{
		q.GET("list", v1.Api.GetQuestions)
		q.POST("", middleware.OperationRecord(), v1.Api.SaveQuestion)
		q.POST("import", middleware.OperationRecord(), v1.Api.ImportQuestions)
		q.DELETE("", middleware.OperationRecord(), v1.Api.DeleteQuestion)
	}
}

// InitTerminal 终端公开路由（机器码白名单鉴权，服务层校验）
func (rg *RouterGroup) InitTerminal(Public *gin.RouterGroup) {
	t := Public.Group("exam/terminal")
	{
		t.POST("register", v1.Api.TerminalRegister)
		t.POST("heartbeat", v1.Api.TerminalHeartbeat)
		t.GET("exams", v1.Api.TerminalExams)
		t.POST("resume", v1.Api.TerminalResume)
		t.POST("save", v1.Api.TerminalSaveAnswers)
		t.POST("submit", v1.Api.TerminalSubmit)
		t.POST("practice", v1.Api.RunPractice)
		t.GET("practiceLogs", v1.Api.GetPracticeLogs)
	}
}
