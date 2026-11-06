// Package router llmops 路由
package router

import (
	"github.com/gin-gonic/gin"

	v1 "github.com/hequan2017/new-ops/server/plugin/llmops/api/v1"
)

type RouterGroup struct{}

var RouterGroupApp = new(RouterGroup)

// Init 注册 llmops 路由（private 组，JWT+casbin 由 initialize.Router 显式挂载）
func (rg *RouterGroup) Init(Router *gin.RouterGroup) {
	base := Router.Group("llmops")
	{
		base.GET("ping", v1.Api.Ping)
	}
	asset := Router.Group("llmops/asset")
	{
		asset.GET("list", v1.Api.GetAssets)
		asset.GET("", v1.Api.GetAsset)
		asset.POST("importDoc", v1.Api.ImportDoc)
		asset.POST("importWeb", v1.Api.ImportWeb)
		asset.POST("importQa", v1.Api.ImportQa)
		asset.POST("importSemi", v1.Api.ImportSemi)
		asset.POST("importMedia", v1.Api.ImportMedia)
		asset.POST("transcribe", v1.Api.TranscribeMedia)
		asset.POST("categorize", v1.Api.CategorizeAsset)
		asset.GET("events", v1.Api.GetAssetEvents)
		asset.DELETE("", v1.Api.DeleteAsset)
	}
	cat := Router.Group("llmops/category")
	{
		cat.POST("", v1.Api.SaveCategory)
		cat.GET("tree", v1.Api.GetCategoryTree)
		cat.DELETE("", v1.Api.DeleteCategory)
	}
	ds := Router.Group("llmops/datasource")
	{
		ds.GET("list", v1.Api.GetDataSources)
		ds.POST("", v1.Api.SaveDataSource)
		ds.POST("test", v1.Api.TestDataSource)
		ds.GET("tables", v1.Api.ListDataSourceTables)
		ds.GET("preview", v1.Api.PreviewDataSourceTable)
		ds.DELETE("", v1.Api.DeleteDataSource)
	}
	ml := Router.Group("llmops/model")
	{
		ml.GET("list", v1.Api.GetModels)
		ml.POST("", v1.Api.SaveModel)
		ml.POST("test", v1.Api.TestModel)
		ml.GET("callLogs", v1.Api.GetModelCallLogs)
		ml.GET("select", v1.Api.SelectModel)
		ml.DELETE("", v1.Api.DeleteModel)
	}
	dep := Router.Group("llmops/deploy")
	{
		dep.GET("list", v1.Api.GetDeploys)
		dep.POST("", v1.Api.SaveDeployPlan)
		dep.POST("execute", v1.Api.ExecuteDeploy)
		dep.POST("destroy", v1.Api.DestroyDeploy)
		dep.DELETE("", v1.Api.DeleteDeploy)
	}
	kb := Router.Group("llmops/kb")
	{
		kb.GET("list", v1.Api.GetKnowledges)
		kb.POST("", v1.Api.SaveKnowledge)
		kb.POST("ingest", v1.Api.IngestKnowledgeDoc)
		kb.GET("docs", v1.Api.GetKnowledgeDocs)
		kb.POST("vector/index", v1.Api.IndexVectorDoc)
		kb.POST("vector/search", v1.Api.SearchVector)
		kb.POST("fulltext/index", v1.Api.IndexFulltextDoc)
		kb.POST("fulltext/search", v1.Api.SearchFulltext)
		kb.POST("qa/search", v1.Api.SearchQa)
		kb.POST("graph/node", v1.Api.SaveGraphNode)
		kb.GET("graph/nodes", v1.Api.GetGraphNodes)
		kb.POST("graph/edge", v1.Api.SaveGraphEdge)
		kb.GET("graph/edges", v1.Api.GetGraphEdges)
		kb.POST("graph/query", v1.Api.GraphMultiHopQuery)
		kb.DELETE("graph/node", v1.Api.DeleteGraphNode)
		kb.DELETE("graph/edge", v1.Api.DeleteGraphEdge)
		kb.POST("summary/generate", v1.Api.GenerateDocSummary)
		kb.GET("summary", v1.Api.GetDocSummary)
		kb.POST("summary/search", v1.Api.SearchSummaries)
	}
	ks := Router.Group("llmops/knowledge")
	{
		ks.POST("search", v1.Api.UnifiedSearch)
		ks.GET("searchLogs", v1.Api.GetSearchLogs)
		kb.DELETE("doc", v1.Api.DeleteKnowledgeDoc)
		kb.DELETE("", v1.Api.DeleteKnowledge)
	}
}
