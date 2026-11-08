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
}
