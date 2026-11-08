// Package router ziqiang 路由
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/middleware"

	v1 "github.com/hequan2017/new-ops/server/plugin/ziqiang/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 ziqiang 路由（private 组，JWT+casbin 由 initialize.Router 显式挂载）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	base := Router.Group("ziqiang")
	{
		base.GET("ping", v1.Api.Ping)
	}
	tc := Router.Group("ziqiang/textCorpus")
	{
		tc.GET("list", v1.Api.GetTextCorpora)
		tc.POST("", middleware.OperationRecord(), v1.Api.SaveTextCorpus)
		tc.POST("import", middleware.OperationRecord(), v1.Api.ImportTextCorpora)
		tc.DELETE("", middleware.OperationRecord(), v1.Api.DeleteTextCorpus)
	}
	cp := Router.Group("ziqiang/comparePair")
	{
		cp.GET("list", v1.Api.GetComparePairs)
		cp.POST("", middleware.OperationRecord(), v1.Api.SaveComparePair)
	}
	ic := Router.Group("ziqiang/imageCorpus")
	{
		ic.GET("list", v1.Api.GetImageCorpora)
		ic.GET("compare", v1.Api.CompareImageHash)
		ic.POST("", middleware.OperationRecord(), v1.Api.SaveImageCorpus)
	}
	Router.Group("ziqiang/termCheck").POST("", v1.Api.CheckTerms)
	Router.Group("ziqiang/dpMask").POST("", v1.Api.DpMaskNumbers)
	Router.Group("ziqiang/legalTokenize").POST("", v1.Api.LegalTokenize)
	Router.Group("ziqiang/plagiarismCheck").POST("", v1.Api.SlidingPlagiarism)
	gEntity := Router.Group("ziqiang/entity")
	{
		gEntity.GET("list", v1.Api.GetGraphEntities)
		gEntity.POST("", middleware.OperationRecord(), v1.Api.SaveGraphEntity)
		gEntity.DELETE("", middleware.OperationRecord(), v1.Api.DeleteGraphEntity)
	}
	gEdge := Router.Group("ziqiang/edge")
	{
		gEdge.POST("", middleware.OperationRecord(), v1.Api.SaveGraphEdge)
		gEdge.DELETE("", middleware.OperationRecord(), v1.Api.DeleteGraphEdge)
	}
	Router.Group("ziqiang/graph").GET("", v1.Api.GetGraph)
	Router.Group("ziqiang/graph").GET("bfs", v1.Api.GraphBFS)
	reg := Router.Group("ziqiang/regulation")
	{
		reg.GET("list", v1.Api.GetRegulations)
		reg.GET("search", v1.Api.SearchRegulations)
		reg.POST("", middleware.OperationRecord(), v1.Api.SaveRegulation)
		reg.DELETE("", middleware.OperationRecord(), v1.Api.DeleteRegulation)
	}
	caseGrp := Router.Group("ziqiang/case")
	{
		caseGrp.GET("list", v1.Api.GetCases)
		caseGrp.POST("", middleware.OperationRecord(), v1.Api.SaveCase)
	}
}
