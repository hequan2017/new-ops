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
		kb.DELETE("doc", v1.Api.DeleteKnowledgeDoc)
		kb.DELETE("", v1.Api.DeleteKnowledge)
	}
	ks := Router.Group("llmops/knowledge")
	{
		ks.POST("search", v1.Api.UnifiedSearch)
		ks.GET("searchLogs", v1.Api.GetSearchLogs)
	}
	tr := Router.Group("llmops/train")
	{
		tr.GET("list", v1.Api.GetTrainJobs)
		tr.POST("", v1.Api.SaveTrainJob)
		tr.POST("transition", v1.Api.TransitionTrainJob)
		tr.POST("dryRun", v1.Api.DryRunTrain)
		tr.POST("start", v1.Api.StartTrainJob)
		tr.POST("stop", v1.Api.StopTrainJob)
		tr.GET("log", v1.Api.GetTrainJobLog)
	}
	ev := Router.Group("llmops/eval")
	{
		ev.GET("list", v1.Api.GetEvalRuns)
		ev.POST("", v1.Api.SaveEvalRun)
		ev.POST("run", v1.Api.RunEval)
		ev.GET("results", v1.Api.GetEvalResults)
	}
	art := Router.Group("llmops/artifact")
	{
		art.GET("list", v1.Api.GetArtifacts)
		art.POST("", v1.Api.RegisterArtifact)
		art.POST("registerModel", v1.Api.RegisterArtifactModel)
		art.DELETE("", v1.Api.DeleteArtifact)
	}
	ag := Router.Group("llmops/agent")
	{
		ag.GET("list", v1.Api.GetAgents)
		ag.POST("", v1.Api.SaveAgent)
		ag.POST("stage", v1.Api.SetAgentStage)
		ag.POST("publish", v1.Api.PublishAgent)
		ag.GET("versions", v1.Api.GetAgentVersions)
		ag.POST("session", v1.Api.CreateAgentSession)
		ag.GET("messages", v1.Api.GetAgentMessages)
		ag.POST("chat", v1.Api.AgentChat)
		ag.POST("memory", v1.Api.SaveAgentMemory)
		ag.GET("memory/list", v1.Api.ListAgentMemories)
		ag.POST("memory/inject", v1.Api.InjectAgentMemories)
		ag.POST("memory/extract", v1.Api.ExtractAgentMemory)
		ag.POST("multiRun", v1.Api.AgentMultiRun)
		ag.POST("debugChat", v1.Api.AgentDebugChat)
		ag.GET("sessions", v1.Api.ListAgentSessions)
		ag.GET("eval/list", v1.Api.GetAgentEvals)
		ag.POST("eval", v1.Api.SaveAgentEval)
		ag.POST("eval/run", v1.Api.RunAgentEval)
		ag.GET("eval/results", v1.Api.GetAgentEvalResults)
		ag.POST("rollback", v1.Api.RollbackAgent)
		ag.GET("stats", v1.Api.GetAgentCallStats)
		ag.GET("monitor", v1.Api.GetAgentMonitor)
		ag.GET("abnormalSessions", v1.Api.GetAgentAbnormalSessions)
		ag.DELETE("memory", v1.Api.DeleteAgentMemory)
		ag.DELETE("", v1.Api.DeleteAgent)
	}
	sk := Router.Group("llmops/skill")
	{
		sk.GET("list", v1.Api.GetSkills)
		sk.POST("", v1.Api.SaveSkill)
		sk.POST("generate", v1.Api.GenerateSkill)
		sk.POST("export", v1.Api.ExportSkillPackage)
		sk.POST("import", v1.Api.ImportSkillPackage)
		sk.POST("call", v1.Api.SkillCall)
		sk.GET("callLogs", v1.Api.GetSkillCallLogs)
		sk.GET("token/list", v1.Api.GetSkillTokens)
		sk.POST("token", v1.Api.CreateSkillToken)
		sk.POST("token/revoke", v1.Api.RevokeSkillToken)
		sk.DELETE("", v1.Api.DeleteSkill)
	}
	mcp := Router.Group("llmops/mcp")
	{
		mcp.GET("list", v1.Api.GetMcpTools)
		mcp.POST("", v1.Api.SaveMcpTool)
		mcp.DELETE("", v1.Api.DeleteMcpTool)
	}
	conn := Router.Group("llmops/connector")
	{
		conn.GET("list", v1.Api.GetConnectors)
		conn.POST("", v1.Api.SaveConnector)
		conn.POST("test", v1.Api.TestConnector)
		conn.DELETE("", v1.Api.DeleteConnector)
	}
	qt := Router.Group("llmops/quota")
	{
		qt.GET("list", v1.Api.GetQuotas)
		qt.POST("", v1.Api.SetQuota)
		qt.POST("consume", v1.Api.ConsumeQuota)
	}
	db := Router.Group("llmops/dashboard")
	{
		db.GET("", v1.Api.GetDashboard)
	}
	dst := Router.Group("llmops/dataset")
	{
		dst.GET("list", v1.Api.GetDatasets)
		dst.POST("", v1.Api.SaveDataset)
		dst.POST("import", v1.Api.ImportDatasetContent)
		dst.GET("export", v1.Api.ExportDataset)
		dst.GET("refs", v1.Api.GetDatasetRefs)
		dst.DELETE("", v1.Api.DeleteDataset)
	}
	pt := Router.Group("llmops/prompt")
	{
		pt.GET("list", v1.Api.GetPrompts)
		pt.POST("", v1.Api.SavePrompt)
		pt.POST("render", v1.Api.RenderPrompt)
		pt.POST("optimize", v1.Api.OptimizePrompts)
		pt.GET("asset/list", v1.Api.GetPromptAssets)
		pt.POST("asset", v1.Api.SavePromptAsset)
		pt.DELETE("asset", v1.Api.DeletePromptAsset)
		pt.GET("versions", v1.Api.GetPromptVersions)
		pt.DELETE("", v1.Api.DeletePrompt)
	}
}
