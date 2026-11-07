// Package initialize llmops 大模型应用开发平台插件初始化（表/API/菜单/授权/路由，全部幂等）
package initialize

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/middleware"
	sysModel "github.com/hequan2017/new-ops/server/model/system"
	pluginUtils "github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
	srvUtils "github.com/hequan2017/new-ops/server/utils"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
	"github.com/hequan2017/new-ops/server/plugin/llmops/router"
)

// Gorm 注册 llmops 数据表（幂等）；表随排期条目逐域新增，新增在此追加
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.LlmopsDataAsset),
		new(model.LlmopsQaPair),
		new(model.LlmopsDataSource),
		new(model.LlmopsCategory),
		new(model.LlmopsAssetEvent),
		new(model.LlmopsModel),
		new(model.LlmopsModelCallLog),
		new(model.LlmopsModelDeploy),
		new(model.LlmopsKnowledge),
		new(model.LlmopsKnowledgeDoc),
		new(model.LlmopsVectorChunk),
		new(model.LlmopsFulltextChunk),
		new(model.LlmopsGraphNode),
		new(model.LlmopsGraphEdge),
		new(model.LlmopsDocSummary),
		new(model.LlmopsSearchLog),
		new(model.LlmopsTrainJob),
		new(model.LlmopsEvalRun),
		new(model.LlmopsEvalResult),
		new(model.LlmopsArtifact),
		new(model.LlmopsAgent),
		new(model.LlmopsAgentVersion),
		new(model.LlmopsAgentSession),
		new(model.LlmopsAgentMessage),
		new(model.LlmopsAgentMemory),
		new(model.LlmopsAgentEval),
		new(model.LlmopsAgentEvalResult),
		new(model.LlmopsAgentCallLog),
		new(model.LlmopsSkill),
		new(model.LlmopsMcpTool),
		new(model.LlmopsConnector),
		new(model.LlmopsSkillToken),
		new(model.LlmopsSkillCallLog),
		new(model.LlmopsPromptTemplate),
		new(model.LlmopsPromptVersion),
		new(model.LlmopsPromptAsset),
		new(model.LlmopsDataset),
		new(model.LlmopsQuota),
	)
	if err != nil {
		err = errors.Wrap(err, "llmops 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册 llmops 路由（private 组显式挂 JWT+casbin）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}

// Api 注册 llmops API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/llmops/ping", Description: "llmops 连通性探测", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/asset/list", Description: "数据资产列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/asset", Description: "数据资产详情", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/asset/importDoc", Description: "文档采集上传", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/asset/importWeb", Description: "网页采集", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/asset/importQa", Description: "问答对采集", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/asset/importSemi", Description: "半结构化编目", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/asset/categorize", Description: "资产挂分类/打标签", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/asset/events", Description: "资产事件流水", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/asset/importMedia", Description: "媒体元数据编目", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/asset/transcribe", Description: "媒体 ASR 转写", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/category", Description: "保存编目分类", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/category/tree", Description: "编目分类树", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/category", Description: "删除编目分类", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/model/list", Description: "模型注册列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/model", Description: "保存模型注册", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/model/test", Description: "模型在线测试", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/model/callLogs", Description: "模型调用日志", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/model/select", Description: "按能力调度模型", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/model", Description: "删除模型注册", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/deploy/list", Description: "模型部署列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/deploy", Description: "保存部署计划", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/deploy/execute", Description: "执行模型部署", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/deploy/destroy", Description: "销毁模型部署", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/deploy", Description: "删除部署计划", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/kb/list", Description: "知识库列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/kb", Description: "保存知识库", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/kb/ingest", Description: "知识文档入库", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/kb/docs", Description: "库内文档列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/kb/vector/index", Description: "文档向量索引", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/kb/vector/search", Description: "向量语义检索", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/kb/fulltext/index", Description: "文档全文索引", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/kb/fulltext/search", Description: "全文检索", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/kb/qa/search", Description: "问答对检索", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/kb/graph/node", Description: "保存图谱节点", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/kb/graph/nodes", Description: "图谱节点列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/kb/graph/edge", Description: "保存图谱边", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/kb/graph/edges", Description: "图谱边列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/kb/graph/query", Description: "图谱多跳推理", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/kb/graph/node", Description: "删除图谱节点", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/kb/graph/edge", Description: "删除图谱边", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/kb/summary/generate", Description: "生成文档摘要", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/kb/summary", Description: "查询文档摘要", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/kb/summary/search", Description: "摘要检索", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/knowledge/search", Description: "统一检索入口", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/knowledge/searchLogs", Description: "检索调用日志", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/train/list", Description: "训练任务列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/train", Description: "保存训练任务", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/train/transition", Description: "训练任务状态流转", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/train/dryRun", Description: "训练任务 dry-run", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/train/start", Description: "启动训练任务", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/train/stop", Description: "停止训练任务", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/train/log", Description: "训练任务日志", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/eval/list", Description: "评估任务列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/eval", Description: "保存评估任务", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/eval/run", Description: "执行模型评估", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/eval/results", Description: "评估明细", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/artifact/list", Description: "产物列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/artifact", Description: "登记产物", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/artifact/registerModel", Description: "产物注册为模型服务", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/artifact", Description: "删除产物", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/agent/list", Description: "智能体列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/agent", Description: "保存智能体", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/agent/stage", Description: "智能体阶段流转", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/agent/publish", Description: "发布智能体", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/agent/versions", Description: "智能体版本列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/agent/session", Description: "创建智能体会话", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/agent/messages", Description: "会话消息", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/agent/chat", Description: "智能体对话", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/agent/memory", Description: "保存长期记忆", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/agent/memory/list", Description: "长期记忆列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/agent/memory/inject", Description: "记忆检索注入", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/agent/memory/extract", Description: "自动记忆抽取", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/agent/multiRun", Description: "多智能体运行", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/agent/debugChat", Description: "智能体调试对话", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/agent/sessions", Description: "智能体会话列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/agent/eval/list", Description: "智能体评测列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/agent/eval", Description: "保存智能体评测", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/agent/eval/run", Description: "执行智能体评测", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/agent/eval/results", Description: "智能体评测明细", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/agent/rollback", Description: "智能体版本回滚", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/agent/stats", Description: "智能体调用统计", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/agent/monitor", Description: "智能体运行监控", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/agent/abnormalSessions", Description: "智能体异常会话", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/skill/list", Description: "技能列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/skill", Description: "保存技能", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/skill/generate", Description: "零代码创建技能", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/skill/export", Description: "技能包导出", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/skill/import", Description: "技能包导入", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/mcp/list", Description: "MCP 工具列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/mcp", Description: "保存 MCP 工具", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/mcp", Description: "删除 MCP 工具", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/connector/list", Description: "连接器列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/connector", Description: "保存连接器", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/connector/test", Description: "连接器连通测试", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/connector", Description: "删除连接器", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/skill", Description: "删除技能", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/skill/call", Description: "技能开放调用", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/skill/callLogs", Description: "技能审计日志", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/skill/token/list", Description: "技能令牌列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/skill/token", Description: "创建技能令牌", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/skill/token/revoke", Description: "吊销技能令牌", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/prompt/list", Description: "Prompt 模板列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/prompt", Description: "保存 Prompt 模板", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/prompt/render", Description: "Prompt 模板渲染", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/prompt/optimize", Description: "Prompt 优化", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/prompt/asset/list", Description: "Prompt 配套资产列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/prompt/asset", Description: "保存 Prompt 配套资产", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/prompt/asset", Description: "删除 Prompt 配套资产", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/dataset/list", Description: "数据集列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/dataset", Description: "保存数据集", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/dataset/import", Description: "数据集导入", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/dataset/export", Description: "数据集导出", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/dataset/refs", Description: "数据集引用计数", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/dashboard", Description: "平台数据看板", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/quota/list", Description: "配额列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/quota", Description: "设置配额", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/quota/consume", Description: "配额扣减", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/dataset", Description: "删除数据集", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/prompt/versions", Description: "Prompt 模板版本", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/prompt", Description: "删除 Prompt 模板", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/agent/memory", Description: "删除长期记忆", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/agent", Description: "删除智能体", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/train", Description: "删除训练任务", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/kb/doc", Description: "删除库内文档", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/kb", Description: "删除知识库", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/datasource/list", Description: "数据源列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/datasource", Description: "保存数据源", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/datasource/test", Description: "数据源连接测试", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/llmops/datasource/tables", Description: "库表清单", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/datasource/preview", Description: "表采样预览", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/llmops/datasource", Description: "删除数据源", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/llmops/asset", Description: "删除数据资产", ApiGroup: "大模型平台", Method: "DELETE"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 llmops 菜单（父级目录，子页面菜单随各 web 条目追加）
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "llmops",
			Name:      "llmops",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      5,
			Meta:      sysModel.Meta{Title: "大模型平台", Icon: "data-line"},
		},
		{
			Path:      "llmopsData",
			Name:      "llmopsData",
			Hidden:    false,
			Component: "plugin/llmops/view/data/index.vue",
			Sort:      1,
			Meta:      sysModel.Meta{Title: "数据管理", Icon: "files"},
		},
		{
			Path:      "llmopsModel",
			Name:      "llmopsModel",
			Hidden:    false,
			Component: "plugin/llmops/view/model/index.vue",
			Sort:      2,
			Meta:      sysModel.Meta{Title: "模型管理", Icon: "cpu"},
		},
		{
			Path:      "llmopsKb",
			Name:      "llmopsKb",
			Hidden:    false,
			Component: "plugin/llmops/view/kb/index.vue",
			Sort:      3,
			Meta:      sysModel.Meta{Title: "知识库", Icon: "collection"},
		},
		{
			Path:      "llmopsTrain",
			Name:      "llmopsTrain",
			Hidden:    false,
			Component: "plugin/llmops/view/train/index.vue",
			Sort:      4,
			Meta:      sysModel.Meta{Title: "训练调优", Icon: "video-play"},
		},
		{
			Path:      "llmopsAgent",
			Name:      "llmopsAgent",
			Hidden:    false,
			Component: "plugin/llmops/view/agent/index.vue",
			Sort:      5,
			Meta:      sysModel.Meta{Title: "智能体", Icon: "chat-dot-round"},
		},
		{
			Path:      "llmopsSkill",
			Name:      "llmopsSkill",
			Hidden:    false,
			Component: "plugin/llmops/view/skill/index.vue",
			Sort:      6,
			Meta:      sysModel.Meta{Title: "技能中心", Icon: "coordinate"},
		},
		{
			Path:      "llmopsPrompt",
			Name:      "llmopsPrompt",
			Hidden:    false,
			Component: "plugin/llmops/view/prompt/index.vue",
			Sort:      7,
			Meta:      sysModel.Meta{Title: "Prompt 工作台", Icon: "edit-pen"},
		},
		{
			Path:      "llmopsDataset",
			Name:      "llmopsDataset",
			Hidden:    false,
			Component: "plugin/llmops/view/dataset/index.vue",
			Sort:      8,
			Meta:      sysModel.Meta{Title: "数据集", Icon: "box"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// PresetModels 预置模型服务种子（幂等 FirstOrCreate；3 个内置定义，待选接入点后一键部署）
func PresetModels(ctx context.Context) {
	presets := []model.LlmopsModelDeploy{
		{
			Name: "preset-vllm-openai", Image: "vllm/vllm-openai:latest", Ports: "8000:8000",
			BaseUrl: "http://127.0.0.1:8000/v1", Status: "pending",
			Notes: "预置·vLLM OpenAI 兼容服务（选择接入点后一键部署）",
		},
		{
			Name: "preset-ollama", Image: "ollama/ollama:latest", Ports: "11434:11434",
			BaseUrl: "http://127.0.0.1:11434", Status: "pending",
			Notes: "预置·Ollama 推理服务（选择接入点后一键部署）",
		},
		{
			Name: "preset-xinference", Image: "xprobe/xinference:latest", Ports: "9997:9997",
			BaseUrl: "http://127.0.0.1:9997/v1", Status: "pending",
			Notes: "预置·Xinference 推理平台（选择接入点后一键部署）",
		},
	}
	for i := range presets {
		var exist model.LlmopsModelDeploy
		err := global.GVA_DB.WithContext(ctx).Where("name = ?", presets[i].Name).First(&exist).Error
		if err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&presets[i]).Error; err != nil {
				global.GVA_LOG.Error(fmt.Sprintf("llmops 预置模型种子失败: %v", err))
			}
		}
	}
}

// BuiltinSkills 内置技能种子（幂等 FirstOrCreate；要求(6) 预置内置技能）
func BuiltinSkills(ctx context.Context) {
	presets := []model.LlmopsSkill{
		{
			Name: "文本摘要", Type: "builtin", Status: "enabled",
			Desc: "将输入文本概括为 3-5 句摘要",
			DefJSON: `{"kind":"prompt","system":"你是摘要助手，只输出摘要本身。","template":"请用 3-5 句话概括以下内容：\n{input}"}`,
		},
		{
			Name: "中英翻译", Type: "builtin", Status: "enabled",
			Desc: "中英互译（可指定目标语言）",
			DefJSON: `{"kind":"prompt","system":"你是专业翻译，只输出译文。","template":"把以下内容翻译成{lang}：\n{input}","params":{"lang":{"desc":"目标语言","default":"英文"}}}`,
		},
		{
			Name: "SQL生成", Type: "builtin", Status: "enabled",
			Desc: "按表结构生成 SQL 查询",
			DefJSON: `{"kind":"prompt","system":"你是 SQL 专家，只输出一条可执行 SQL，不要解释。","template":"表结构：\n{schema}\n\n需求：{input}","params":{"schema":{"desc":"建表语句或字段说明","default":"(用户补充表结构)"}}}`,
		},
		{
			Name: "关键词提取", Type: "builtin", Status: "enabled",
			Desc: "提取文本关键词，输出 JSON 数组",
			DefJSON: `{"kind":"prompt","system":"只输出 JSON 字符串数组，不要其它内容。","template":"提取以下文本的关键词（最多 10 个）：\n{input}"}`,
		},
		{
			Name: "表格问答", Type: "builtin", Status: "enabled",
			Desc: "基于 markdown 表格回答问题",
			DefJSON: `{"kind":"prompt","system":"根据表格回答问题，答案简洁并注明依据行。","template":"表格：\n{table}\n\n问题：{input}","params":{"table":{"desc":"markdown 表格","default":"(用户补充表格)"}}}`,
		},
	}
	for i := range presets {
		var exist model.LlmopsSkill
		if err := global.GVA_DB.WithContext(ctx).Where("name = ?", presets[i].Name).First(&exist).Error; err != nil {
			if err := global.GVA_DB.WithContext(ctx).Create(&presets[i]).Error; err != nil {
				global.GVA_LOG.Error(fmt.Sprintf("llmops 内置技能种子失败: %v", err))
			}
		}
	}
}

// Casbin 注册 llmops 策略（幂等）：仅 888 全量（平台管理面）
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/llmops/ping", "GET"},
		{"/llmops/asset/list", "GET"},
		{"/llmops/asset", "GET"},
		{"/llmops/asset/importDoc", "POST"},
		{"/llmops/asset/importWeb", "POST"},
		{"/llmops/asset/importQa", "POST"},
		{"/llmops/asset/importSemi", "POST"},
		{"/llmops/asset/categorize", "POST"},
		{"/llmops/asset/events", "GET"},
		{"/llmops/asset/importMedia", "POST"},
		{"/llmops/asset/transcribe", "POST"},
		{"/llmops/category", "POST"},
		{"/llmops/category/tree", "GET"},
		{"/llmops/category", "DELETE"},
		{"/llmops/model/list", "GET"},
		{"/llmops/model", "POST"},
		{"/llmops/model/test", "POST"},
		{"/llmops/model/callLogs", "GET"},
		{"/llmops/model/select", "GET"},
		{"/llmops/model", "DELETE"},
		{"/llmops/deploy/list", "GET"},
		{"/llmops/deploy", "POST"},
		{"/llmops/deploy/execute", "POST"},
		{"/llmops/deploy/destroy", "POST"},
		{"/llmops/deploy", "DELETE"},
		{"/llmops/kb/list", "GET"},
		{"/llmops/kb", "POST"},
		{"/llmops/kb/ingest", "POST"},
		{"/llmops/kb/docs", "GET"},
		{"/llmops/kb/vector/index", "POST"},
		{"/llmops/kb/vector/search", "POST"},
		{"/llmops/kb/fulltext/index", "POST"},
		{"/llmops/kb/fulltext/search", "POST"},
		{"/llmops/kb/qa/search", "POST"},
		{"/llmops/kb/graph/node", "POST"},
		{"/llmops/kb/graph/nodes", "GET"},
		{"/llmops/kb/graph/edge", "POST"},
		{"/llmops/kb/graph/edges", "GET"},
		{"/llmops/kb/graph/query", "POST"},
		{"/llmops/kb/graph/node", "DELETE"},
		{"/llmops/kb/graph/edge", "DELETE"},
		{"/llmops/kb/summary/generate", "POST"},
		{"/llmops/kb/summary", "GET"},
		{"/llmops/kb/summary/search", "POST"},
		{"/llmops/knowledge/search", "POST"},
		{"/llmops/knowledge/searchLogs", "GET"},
		{"/llmops/train/list", "GET"},
		{"/llmops/train", "POST"},
		{"/llmops/train/transition", "POST"},
		{"/llmops/train/dryRun", "POST"},
		{"/llmops/train/start", "POST"},
		{"/llmops/train/stop", "POST"},
		{"/llmops/train/log", "GET"},
		{"/llmops/eval/list", "GET"},
		{"/llmops/eval", "POST"},
		{"/llmops/eval/run", "POST"},
		{"/llmops/eval/results", "GET"},
		{"/llmops/artifact/list", "GET"},
		{"/llmops/artifact", "POST"},
		{"/llmops/artifact/registerModel", "POST"},
		{"/llmops/artifact", "DELETE"},
		{"/llmops/agent/list", "GET"},
		{"/llmops/agent", "POST"},
		{"/llmops/agent/stage", "POST"},
		{"/llmops/agent/publish", "POST"},
		{"/llmops/agent/versions", "GET"},
		{"/llmops/agent/session", "POST"},
		{"/llmops/agent/messages", "GET"},
		{"/llmops/agent/chat", "POST"},
		{"/llmops/agent/memory", "POST"},
		{"/llmops/agent/memory/list", "GET"},
		{"/llmops/agent/memory/inject", "POST"},
		{"/llmops/agent/memory/extract", "POST"},
		{"/llmops/agent/multiRun", "POST"},
		{"/llmops/agent/debugChat", "POST"},
		{"/llmops/agent/sessions", "GET"},
		{"/llmops/agent/eval/list", "GET"},
		{"/llmops/agent/eval", "POST"},
		{"/llmops/agent/eval/run", "POST"},
		{"/llmops/agent/eval/results", "GET"},
		{"/llmops/agent/rollback", "POST"},
		{"/llmops/agent/stats", "GET"},
		{"/llmops/agent/monitor", "GET"},
		{"/llmops/agent/abnormalSessions", "GET"},
		{"/llmops/skill/list", "GET"},
		{"/llmops/skill", "POST"},
		{"/llmops/skill/generate", "POST"},
		{"/llmops/skill/export", "POST"},
		{"/llmops/skill/import", "POST"},
		{"/llmops/mcp/list", "GET"},
		{"/llmops/mcp", "POST"},
		{"/llmops/mcp", "DELETE"},
		{"/llmops/connector/list", "GET"},
		{"/llmops/connector", "POST"},
		{"/llmops/connector/test", "POST"},
		{"/llmops/connector", "DELETE"},
		{"/llmops/skill", "DELETE"},
		{"/llmops/skill/call", "POST"},
		{"/llmops/skill/callLogs", "GET"},
		{"/llmops/skill/token/list", "GET"},
		{"/llmops/skill/token", "POST"},
		{"/llmops/skill/token/revoke", "POST"},
		{"/llmops/prompt/list", "GET"},
		{"/llmops/prompt", "POST"},
		{"/llmops/prompt/render", "POST"},
		{"/llmops/prompt/optimize", "POST"},
		{"/llmops/prompt/asset/list", "GET"},
		{"/llmops/prompt/asset", "POST"},
		{"/llmops/prompt/asset", "DELETE"},
		{"/llmops/dataset/list", "GET"},
		{"/llmops/dataset", "POST"},
		{"/llmops/dataset/import", "POST"},
		{"/llmops/dataset/export", "GET"},
		{"/llmops/dataset/refs", "GET"},
		{"/llmops/dashboard", "GET"},
		{"/llmops/quota/list", "GET"},
		{"/llmops/quota", "POST"},
		{"/llmops/quota/consume", "POST"},
		{"/llmops/dataset", "DELETE"},
		{"/llmops/prompt/versions", "GET"},
		{"/llmops/prompt", "DELETE"},
		{"/llmops/agent/memory", "DELETE"},
		{"/llmops/agent", "DELETE"},
		{"/llmops/train", "DELETE"},
		{"/llmops/kb/doc", "DELETE"},
		{"/llmops/kb", "DELETE"},
		{"/llmops/datasource/list", "GET"},
		{"/llmops/datasource", "POST"},
		{"/llmops/datasource/test", "POST"},
		{"/llmops/datasource/tables", "GET"},
		{"/llmops/datasource/preview", "GET"},
		{"/llmops/datasource", "DELETE"},
		{"/llmops/asset", "DELETE"},
	} {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			global.GVA_LOG.Error("llmops casbin 添加失败: " + err.Error())
		}
	}
}
