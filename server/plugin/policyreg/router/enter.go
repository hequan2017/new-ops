// Package router policyreg 路由
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/middleware"

	v1 "github.com/hequan2017/new-ops/server/plugin/policyreg/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 policyreg 路由（private 组，JWT+casbin 由 initialize.Router 显式挂载）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	base := Router.Group("policyreg")
	{
		base.GET("ping", v1.Api.Ping)
	}
	doc := Router.Group("policyreg/doc")
	{
		doc.GET("list", v1.Api.GetPolicyDocs)
		doc.GET("", v1.Api.GetPolicyDoc)
		doc.POST("", middleware.OperationRecord(), v1.Api.SavePolicyDoc)
	}
	Router.Group("policyreg/clean").POST("", middleware.OperationRecord(), v1.Api.CleanPolicyDoc)
	term := Router.Group("policyreg/term")
	{
		term.POST("tag", v1.Api.TagPolicyTerms)
		term.GET("tag/list", v1.Api.GetPolicyTermTags)
	}
	ent := Router.Group("policyreg/entity")
	{
		ent.POST("extract", v1.Api.ExtractPolicyEntities)
		ent.GET("list", v1.Api.GetPolicyEntities)
	}
	lb := Router.Group("policyreg/label")
	{
		lb.POST("pre", v1.Api.BuildPolicyLabels)
		lb.POST("token", v1.Api.UpdatePolicyTokenLabel)
		lb.GET("list", v1.Api.GetPolicyTokenLabels)
		lb.GET("relations", v1.Api.GetPolicyRelationLabels)
	}
}
