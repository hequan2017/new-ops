package router

import "github.com/gin-gonic/gin"

type AutoCodeRouter struct{}

func (s *AutoCodeRouter) InitAutoCodeRouter(Router *gin.RouterGroup, _ *gin.RouterGroup) {
	autoCodeRouter := Router.Group("autoCode")
	{
		autoCodeRouter.GET("getDB", autoCodeApi.GetDB)
		autoCodeRouter.GET("getTables", autoCodeApi.GetTables)
		autoCodeRouter.GET("getColumn", autoCodeApi.GetColumn)
	}
	{
		autoCodeRouter.POST("preview", autoCodeTemplateApi.Preview)
		autoCodeRouter.POST("createTemp", autoCodeTemplateApi.Create)
		autoCodeRouter.POST("addFunc", autoCodeTemplateApi.AddFunc)
	}
	{
		autoCodeRouter.POST("mcp", autoCodeTemplateApi.MCP)
		autoCodeRouter.POST("mcpStatus", autoCodeTemplateApi.MCPStatus)
		autoCodeRouter.POST("mcpStart", autoCodeTemplateApi.MCPStart)
		autoCodeRouter.POST("mcpStop", autoCodeTemplateApi.MCPStop)
		autoCodeRouter.POST("mcpList", autoCodeTemplateApi.MCPList)
		autoCodeRouter.POST("mcpRoutes", autoCodeTemplateApi.MCPRoutes)
		autoCodeRouter.POST("mcpTest", autoCodeTemplateApi.MCPTest)
	}
	{
		autoCodeRouter.POST("getPackage", autoCodePackageApi.All)
		autoCodeRouter.POST("delPackage", autoCodePackageApi.Delete)
		autoCodeRouter.POST("createPackage", autoCodePackageApi.Create)
		autoCodeRouter.POST("saveAIWorkflowSession", aiWorkflowSessionApi.Save)
		autoCodeRouter.POST("getAIWorkflowSessionList", aiWorkflowSessionApi.GetList)
		autoCodeRouter.POST("getAIWorkflowSessionDetail", aiWorkflowSessionApi.GetDetail)
		autoCodeRouter.POST("deleteAIWorkflowSession", aiWorkflowSessionApi.Delete)
		autoCodeRouter.POST("dumpAIWorkflowMarkdown", aiWorkflowSessionApi.DumpMarkdown)
	}
	{
		autoCodeRouter.GET("getTemplates", autoCodePackageApi.Templates)
	}
	{
		autoCodeRouter.POST("pubPlug", autoCodePluginApi.Packaged)
		autoCodeRouter.POST("installPlugin", autoCodePluginApi.Install)
		autoCodeRouter.POST("removePlugin", autoCodePluginApi.Remove)
		autoCodeRouter.GET("getPluginList", autoCodePluginApi.GetPluginList)
	}
	{
		// 与 router/system 同步收敛：LLM 代理与插件初始化原挂 public 组（无 JWT/casbin），
		// 是未鉴权攻击面的第二个注册点（巡检二轮 P0-1 只收敛了 system 侧，此处漏网）
		autoCodeRouter.POST("llmAuto", autoCodeApi.LLMAuto)
		autoCodeRouter.POST("llmAutoSSE", autoCodeApi.LLMAutoSSE)
		autoCodeRouter.POST("initMenu", autoCodePluginApi.InitMenu)
		autoCodeRouter.POST("initAPI", autoCodePluginApi.InitAPI)
		autoCodeRouter.POST("initDictionary", autoCodePluginApi.InitDictionary)
	}
}
