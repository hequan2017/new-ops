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
