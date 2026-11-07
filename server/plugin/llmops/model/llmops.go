// Package model llmops 大模型应用开发平台：数据资产域
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// LlmopsDataAsset 多模态数据资产（要求(1) 数据管理与知识治理）
// dataType：doc/image/audio/video/webpage/qa/structured/semistructured
// status：draft(采集) → parsed(解析) → cataloged(编目可用) → offline(下线)
type LlmopsDataAsset struct {
	global.GVA_MODEL
	Name       string `json:"name" gorm:"comment:资产名称"`
	DataType   string `json:"dataType" gorm:"type:varchar(32);index;comment:数据类型"`
	Source     string `json:"source" gorm:"type:varchar(32);comment:来源 upload/url/datasource/import"`
	Status     string `json:"status" gorm:"type:varchar(16);index;default:draft;comment:状态 draft/parsed/cataloged/offline"`
	RawText    string `json:"rawText" gorm:"type:longtext;comment:解析后正文"`
	MetaJSON   string `json:"metaJSON" gorm:"type:text;comment:编目元数据JSON"`
	Tags       string `json:"tags" gorm:"type:varchar(255);comment:标签(逗号分隔)"`
	CategoryId uint   `json:"categoryId" gorm:"index;comment:分类ID"`
	Size       int64  `json:"size" gorm:"comment:原始大小(字节)"`
}

func (LlmopsDataAsset) TableName() string { return "llmops_data_asset" }

// LlmopsQaPair 问答对明细（要求(1) 问答对数据类型；C05 问答对库检索复用）
type LlmopsQaPair struct {
	global.GVA_MODEL
	AssetId  uint   `json:"assetId" gorm:"index;comment:来源资产ID"`
	Question string `json:"question" gorm:"type:varchar(512);index;comment:问题"`
	Answer   string `json:"answer" gorm:"type:text;comment:答案"`
}

func (LlmopsQaPair) TableName() string { return "llmops_qa_pair" }

// LlmopsDataSource 结构化数据源接入（要求(1) 结构化数据源直接接入）
// 密码 AES-GCM 密文落库（json:"-" 永不序列化回显），对齐 aiops 供应商口径
type LlmopsDataSource struct {
	global.GVA_MODEL
	Name        string `json:"name" gorm:"comment:数据源名称;unique"`
	Type        string `json:"type" gorm:"type:varchar(16);index;comment:类型 mysql/postgres"`
	Host        string `json:"host" gorm:"comment:主机"`
	Port        int    `json:"port" gorm:"comment:端口"`
	Database    string `json:"database" gorm:"comment:库名"`
	User        string `json:"user" gorm:"comment:用户名"`
	PasswordEnc string `json:"-" gorm:"type:text;comment:密码密文(AES-GCM)"`
	Notes       string `json:"notes" gorm:"type:text;comment:备注"`
}

func (LlmopsDataSource) TableName() string { return "llmops_data_source" }

// LlmopsCategory 编目分类目录树（parentId 父引用，0 为根）
type LlmopsCategory struct {
	global.GVA_MODEL
	Name     string `json:"name" gorm:"comment:分类名"`
	ParentId uint   `json:"parentId" gorm:"index;default:0;comment:父分类ID(0=根)"`
	Sort     int    `json:"sort" gorm:"default:0;comment:排序"`
	Remark   string `json:"remark" gorm:"comment:备注"`
}

func (LlmopsCategory) TableName() string { return "llmops_category" }

// LlmopsAssetEvent 资产生命周期事件流水（采集/解析/编目/删除自动留痕，要求(1) 全生命周期管理）
type LlmopsAssetEvent struct {
	global.GVA_MODEL
	AssetId uint   `json:"assetId" gorm:"index;comment:资产ID"`
	Action  string `json:"action" gorm:"type:varchar(32);comment:动作"`
	Detail  string `json:"detail" gorm:"type:text;comment:详情"`
}

func (LlmopsAssetEvent) TableName() string { return "llmops_asset_event" }

// LlmopsModel 模型注册表（要求(3) 模型接入与管理：统一接入/纳管/能力标签调度）
// capability 逗号分隔白名单：chat,embed,rerank；DefaultFor 该模型担任默认调度的能力
type LlmopsModel struct {
	global.GVA_MODEL
	Name       string `json:"name" gorm:"comment:模型注册名;unique"`
	Provider   string `json:"provider" gorm:"comment:供应商"`
	BaseUrl    string `json:"baseUrl" gorm:"comment:API基础地址(含/v1)"`
	ApiKeyEnc  string `json:"-" gorm:"type:text;comment:APIKey密文(AES-GCM)"`
	Protocol   string `json:"protocol" gorm:"type:varchar(16);default:openai;comment:协议 openai"`
	Capability string `json:"capability" gorm:"type:varchar(64);comment:能力标签(逗号分隔)"`
	CtxLen     int    `json:"ctxLen" gorm:"default:8192;comment:上下文长度"`
	Status     string `json:"status" gorm:"type:varchar(16);default:enabled;index;comment:enabled/disabled"`
	DefaultFor string `json:"defaultFor" gorm:"type:varchar(16);comment:担任默认调度的能力"`
	Notes      string `json:"notes" gorm:"type:text;comment:备注"`
}

func (LlmopsModel) TableName() string { return "llmops_model" }

// LlmopsModelCallLog 模型调用日志（要求(3) 在线测试留痕与调度统计基础）
type LlmopsModelCallLog struct {
	global.GVA_MODEL
	ModelId          uint   `json:"modelId" gorm:"index;comment:模型ID"`
	ModelName        string `json:"modelName" gorm:"comment:模型名"`
	Protocol         string `json:"protocol" gorm:"type:varchar(16);index;comment:协议 chat/embed/rerank"`
	LatencyMs        int64  `json:"latencyMs" gorm:"comment:耗时(毫秒)"`
	PromptTokens     int    `json:"promptTokens" gorm:"default:0;comment:输入token"`
	CompletionTokens int    `json:"completionTokens" gorm:"default:0;comment:输出token"`
	TotalTokens      int    `json:"totalTokens" gorm:"default:0;comment:总token"`
	Ok               bool   `json:"ok" gorm:"index;comment:是否成功"`
	Error            string `json:"error" gorm:"type:varchar(500);comment:错误摘要"`
}

func (LlmopsModelCallLog) TableName() string { return "llmops_model_call_log" }

// LlmopsModelDeploy 模型服务一键部署（要求(3) 模型服务一键部署；容器经 container 插件创建）
type LlmopsModelDeploy struct {
	global.GVA_MODEL
	Name        string `json:"name" gorm:"comment:部署名;unique"`
	EndpointId  uint   `json:"endpointId" gorm:"comment:容器接入点ID"`
	Image       string `json:"image" gorm:"comment:镜像"`
	ContainerId string `json:"containerId" gorm:"comment:容器ID"`
	Ports       string `json:"ports" gorm:"type:varchar(255);comment:端口映射(逗号分隔 host:ct)"`
	Envs        string `json:"envs" gorm:"type:text;comment:环境变量(逗号分隔 K=V)"`
	BaseUrl     string `json:"baseUrl" gorm:"comment:部署后服务地址(健康探测)"`
	Status      string `json:"status" gorm:"type:varchar(16);default:pending;index;comment:pending/deploying/running/stopped/failed"`
	ErrMsg      string `json:"errMsg" gorm:"type:varchar(500);comment:最近错误"`
	Notes       string `json:"notes" gorm:"type:text;comment:备注"`
}

func (LlmopsModelDeploy) TableName() string { return "llmops_model_deploy" }

// LlmopsKnowledge 知识库（要求(2) 四类库：vector 向量语义库/fulltext 全文检索库/graph 知识图谱库/qa 问答对库）
type LlmopsKnowledge struct {
	global.GVA_MODEL
	Name         string `json:"name" gorm:"comment:知识库名;unique"`
	Type         string `json:"type" gorm:"type:varchar(16);index;comment:类型 vector/fulltext/graph/qa"`
	Desc         string `json:"desc" gorm:"type:varchar(500);comment:描述"`
	Status       string `json:"status" gorm:"type:varchar(16);default:active;comment:active/disabled"`
	EmbedModelId uint   `json:"embedModelId" gorm:"default:0;comment:默认嵌入模型ID(vector 库)"`
}

func (LlmopsKnowledge) TableName() string { return "llmops_knowledge" }

// LlmopsKnowledgeDoc 知识库文档（原文留档；切块/向量化由检索条目回填 chunkCount）
type LlmopsKnowledgeDoc struct {
	global.GVA_MODEL
	KbId       uint   `json:"kbId" gorm:"index;comment:知识库ID"`
	AssetId    uint   `json:"assetId" gorm:"default:0;comment:来源资产ID(0=内联文本)"`
	Title      string `json:"title" gorm:"comment:标题"`
	Content    string `json:"content" gorm:"type:longtext;comment:原文"`
	Status     string `json:"status" gorm:"type:varchar(16);default:ready;comment:ready/indexing"`
	ChunkCount int    `json:"chunkCount" gorm:"default:0;comment:切块数"`
}

func (LlmopsKnowledgeDoc) TableName() string { return "llmops_knowledge_doc" }

// LlmopsVectorChunk 向量库分块（vectorJSON 存 float 数组 JSON；余弦检索内存计算，万级块内可接受）
type LlmopsVectorChunk struct {
	global.GVA_MODEL
	KbId       uint   `json:"kbId" gorm:"index;comment:知识库ID"`
	DocId      uint   `json:"docId" gorm:"index;comment:文档ID"`
	ChunkIndex int    `json:"chunkIndex" gorm:"comment:块序号"`
	Content    string `json:"content" gorm:"type:text;comment:块文本"`
	VectorJSON string `json:"-" gorm:"type:mediumtext;comment:向量JSON"`
	Dim        int    `json:"dim" gorm:"comment:维度"`
	Source     string `json:"source" gorm:"type:varchar(32);comment:来源 embed:<模型ID>/hash-fallback"`
}

func (LlmopsVectorChunk) TableName() string { return "llmops_vector_chunk" }

// LlmopsFulltextChunk 全文库分块（检索时内存 bigram 统计 + 简化 BM25，万级块内可接受）
type LlmopsFulltextChunk struct {
	global.GVA_MODEL
	KbId       uint   `json:"kbId" gorm:"index;comment:知识库ID"`
	DocId      uint   `json:"docId" gorm:"index;comment:文档ID"`
	ChunkIndex int    `json:"chunkIndex" gorm:"comment:块序号"`
	Content    string `json:"content" gorm:"type:text;comment:块文本"`
}

func (LlmopsFulltextChunk) TableName() string { return "llmops_fulltext_chunk" }

// LlmopsGraphNode 知识图谱节点（要求(2) 知识图谱库多跳推理）
type LlmopsGraphNode struct {
	global.GVA_MODEL
	KbId  uint   `json:"kbId" gorm:"index;comment:知识库ID"`
	Name  string `json:"name" gorm:"type:varchar(128);comment:节点名"`
	Type  string `json:"type" gorm:"type:varchar(32);comment:节点类型(实体/概念)"`
	Props string `json:"props" gorm:"type:text;comment:属性JSON"`
}

func (LlmopsGraphNode) TableName() string { return "llmops_graph_node" }

// LlmopsGraphEdge 知识图谱边（有向）
type LlmopsGraphEdge struct {
	global.GVA_MODEL
	KbId       uint    `json:"kbId" gorm:"index;comment:知识库ID"`
	FromNodeId uint    `json:"fromNodeId" gorm:"index;comment:起点节点ID"`
	ToNodeId   uint    `json:"toNodeId" gorm:"index;comment:终点节点ID"`
	Relation   string  `json:"relation" gorm:"type:varchar(64);comment:关系名"`
	Weight     float64 `json:"weight" gorm:"default:1;comment:权重"`
}

func (LlmopsGraphEdge) TableName() string { return "llmops_graph_edge" }

// LlmopsDocSummary 文档摘要（异步生成，要求(2) 摘要检索多路召回）
type LlmopsDocSummary struct {
	global.GVA_MODEL
	KbId    uint   `json:"kbId" gorm:"index;comment:知识库ID"`
	DocId   uint   `json:"docId" gorm:"index;comment:文档ID"`
	Summary string `json:"summary" gorm:"type:text;comment:摘要内容"`
	Status  string `json:"status" gorm:"type:varchar(16);default:pending;index;comment:pending/running/done/failed"`
	ModelId uint   `json:"modelId" gorm:"comment:生成用 chat 模型ID"`
	ErrMsg  string `json:"errMsg" gorm:"type:varchar(500);comment:错误"`
}

func (LlmopsDocSummary) TableName() string { return "llmops_doc_summary" }

// LlmopsSearchLog 统一检索调用日志（要求(2) 调用留痕）
type LlmopsSearchLog struct {
	global.GVA_MODEL
	KbIds         string `json:"kbIds" gorm:"type:varchar(255);comment:库ID列表"`
	Query         string `json:"query" gorm:"type:varchar(512);comment:查询"`
	Mode          string `json:"mode" gorm:"type:varchar(16);comment:hybrid/单路由"`
	Routes        string `json:"routes" gorm:"type:varchar(128);comment:实际路由"`
	TopK          int    `json:"topK" gorm:"comment:topK"`
	HitCount      int    `json:"hitCount" gorm:"comment:命中数"`
	LatencyMs     int64  `json:"latencyMs" gorm:"comment:耗时毫秒"`
	RerankModelId uint   `json:"rerankModelId" gorm:"comment:精排模型"`
	RouteReport   string `json:"routeReport" gorm:"type:varchar(500);comment:各路报告"`
	ErrMsg        string `json:"errMsg" gorm:"type:varchar(500);comment:错误"`
}

func (LlmopsSearchLog) TableName() string { return "llmops_search_log" }

// LlmopsTrainJob 训练/调优任务（要求(4)：SFT 全参/LoRA、RLHF 奖励/强化、评估、压缩）
type LlmopsTrainJob struct {
	global.GVA_MODEL
	Name      string `json:"name" gorm:"comment:任务名;unique"`
	Type      string `json:"type" gorm:"type:varchar(16);index;comment:sft_full/sft_lora/reward/rl/eval/compress"`
	DatasetId uint   `json:"datasetId" gorm:"default:0;comment:数据集ID(K 阶段联动校验)"`
	BaseModel string `json:"baseModel" gorm:"comment:基座模型"`
	HyperJSON string `json:"hyperJSON" gorm:"type:text;comment:超参JSON"`
	PreJobId  uint   `json:"preJobId" gorm:"default:0;comment:前置任务ID(RLHF 两阶段依赖)"`
	Status    string `json:"status" gorm:"type:varchar(16);default:draft;index;comment:draft/queued/running/succeeded/failed/stopped"`
	LogText   string `json:"logText" gorm:"type:mediumtext;comment:执行日志"`
	ErrMsg    string `json:"errMsg" gorm:"type:varchar(500);comment:错误"`
}

func (LlmopsTrainJob) TableName() string { return "llmops_train_job" }

// LlmopsEvalRun 模型评估任务（要求(4) 模型评估；样本来源 SamplesJSON，K 阶段接数据集）
type LlmopsEvalRun struct {
	global.GVA_MODEL
	Name        string `json:"name" gorm:"comment:评估名;unique"`
	ModelId     uint   `json:"modelId" gorm:"comment:被测 chat 模型ID"`
	DatasetId   uint   `json:"datasetId" gorm:"default:0;comment:评测集数据集ID(K 阶段联动)"`
	SamplesJSON string `json:"samplesJSON" gorm:"type:mediumtext;comment:样本JSON[{question,expected}]"`
	MetricJSON  string `json:"metricJSON" gorm:"type:text;comment:指标定义JSON"`
	Status      string `json:"status" gorm:"type:varchar(16);default:pending;index;comment:pending/running/done/failed"`
	SampleCount int    `json:"sampleCount" gorm:"default:0;comment:样本数"`
	PassCount   int    `json:"passCount" gorm:"default:0;comment:通过数"`
	ErrMsg      string `json:"errMsg" gorm:"type:varchar(500);comment:错误"`
}

func (LlmopsEvalRun) TableName() string { return "llmops_eval_run" }

// LlmopsEvalResult 评估明细
type LlmopsEvalResult struct {
	global.GVA_MODEL
	RunId    uint    `json:"runId" gorm:"index;comment:评估任务ID"`
	Idx      int     `json:"idx" gorm:"comment:样本序号"`
	Question string  `json:"question" gorm:"type:varchar(1024);comment:问题"`
	Expected string  `json:"expected" gorm:"type:text;comment:期望答案"`
	Actual   string  `json:"actual" gorm:"type:text;comment:实际输出"`
	Score    float64 `json:"score" gorm:"comment:得分0-1"`
	Pass     bool    `json:"pass" gorm:"comment:是否通过"`
}

func (LlmopsEvalResult) TableName() string { return "llmops_eval_result" }

// LlmopsArtifact 训练/压缩产物（D07 预测服务注册的对接面）
type LlmopsArtifact struct {
	global.GVA_MODEL
	Name    string `json:"name" gorm:"comment:产物名"`
	JobId   uint   `json:"jobId" gorm:"index;comment:来源任务ID"`
	JobType string `json:"jobType" gorm:"type:varchar(16);comment:来源任务类型"`
	Path    string `json:"path" gorm:"comment:产物路径"`
	Method  string `json:"method" gorm:"type:varchar(16);comment:压缩方法(quantize/prune/distill)"`
	Notes   string `json:"notes" gorm:"type:text;comment:备注"`
}

func (LlmopsArtifact) TableName() string { return "llmops_artifact" }

// LlmopsAgent 智能体（要求(5) 全生命周期六阶段：创建/调测/评测/发布/运营/监控）
type LlmopsAgent struct {
	global.GVA_MODEL
	Name         string `json:"name" gorm:"comment:智能体名;unique"`
	Mode         string `json:"mode" gorm:"type:varchar(16);index;comment:qa/interactive/workflow/multi/agentic"`
	Desc         string `json:"desc" gorm:"type:varchar(500);comment:描述"`
	SysPrompt    string `json:"sysPrompt" gorm:"type:text;comment:系统提示词"`
	ModelId      uint   `json:"modelId" gorm:"comment:主模型ID"`
	KnowledgeIds string `json:"knowledgeIds" gorm:"type:varchar(255);comment:挂接知识库ID(逗号分隔)"`
	MaxTurns     int    `json:"maxTurns" gorm:"default:5;comment:工具/反思轮次上限"`
	ConfigJSON   string `json:"configJSON" gorm:"type:text;comment:模式专属配置JSON(工作流节点/子智能体等)"`
	Status       string `json:"status" gorm:"type:varchar(16);default:draft;index;comment:draft/debug/eval/published/operating/monitor"`
}

func (LlmopsAgent) TableName() string { return "llmops_agent" }

// LlmopsAgentVersion 智能体发布版本快照（发布/回滚依据）
type LlmopsAgentVersion struct {
	global.GVA_MODEL
	AgentId  uint   `json:"agentId" gorm:"index;comment:智能体ID"`
	Version  int    `json:"version" gorm:"comment:版本号"`
	Snapshot string `json:"snapshot" gorm:"type:text;comment:发布时配置快照JSON"`
	Remark   string `json:"remark" gorm:"type:varchar(255);comment:发布说明"`
}

func (LlmopsAgentVersion) TableName() string { return "llmops_agent_version" }

// LlmopsAgentSession 智能体会话
type LlmopsAgentSession struct {
	global.GVA_MODEL
	AgentId uint   `json:"agentId" gorm:"index;comment:智能体ID"`
	Title   string `json:"title" gorm:"type:varchar(255);comment:会话标题"`
	UserKey string `json:"userKey" gorm:"type:varchar(64);comment:用户标识"`
}

func (LlmopsAgentSession) TableName() string { return "llmops_agent_session" }

// LlmopsAgentMessage 会话消息（assistant 可携带任务拆解计划与工具回路 trace）
type LlmopsAgentMessage struct {
	global.GVA_MODEL
	SessionId uint   `json:"sessionId" gorm:"index;comment:会话ID"`
	Role      string `json:"role" gorm:"type:varchar(16);comment:user/assistant/system"`
	Content   string `json:"content" gorm:"type:mediumtext;comment:内容"`
	PlanJSON  string `json:"planJSON" gorm:"type:text;comment:任务拆解计划JSON"`
	TraceJSON string `json:"traceJSON" gorm:"type:text;comment:工具回路traceJSON"`
}

func (LlmopsAgentMessage) TableName() string { return "llmops_agent_message" }

// LlmopsAgentMemory 智能体长期记忆条目（要求(5) 长期记忆）
type LlmopsAgentMemory struct {
	global.GVA_MODEL
	AgentId uint    `json:"agentId" gorm:"index;comment:智能体ID"`
	UserKey string  `json:"userKey" gorm:"type:varchar(64);index;comment:用户标识"`
	Content string  `json:"content" gorm:"type:varchar(1024);comment:记忆内容"`
	Source  string  `json:"source" gorm:"type:varchar(16);comment:manual/auto"`
	Weight  float64 `json:"weight" gorm:"default:1;comment:权重"`
}

func (LlmopsAgentMemory) TableName() string { return "llmops_agent_memory" }

// LlmopsAgentEval 智能体评测任务（要求(5) 评测阶段）
type LlmopsAgentEval struct {
	global.GVA_MODEL
	Name         string `json:"name" gorm:"comment:评测名;unique"`
	AgentId      uint   `json:"agentId" gorm:"index;comment:被测智能体ID"`
	SamplesJSON  string `json:"samplesJSON" gorm:"type:mediumtext;comment:样本[{question,expected}]"`
	ScorerType   string `json:"scorerType" gorm:"type:varchar(16);default:contains;comment:contains/keyword/llm"`
	JudgeModelId uint   `json:"judgeModelId" gorm:"default:0;comment:LLM 评分模型(llm 型必填)"`
	Keywords     string `json:"keywords" gorm:"type:varchar(500);comment:keyword 词表(逗号分隔)"`
	Status       string `json:"status" gorm:"type:varchar(16);default:pending;index;comment:pending/running/done/failed"`
	SampleCount  int    `json:"sampleCount" gorm:"default:0;comment:样本数"`
	PassCount    int    `json:"passCount" gorm:"default:0;comment:通过数"`
	ErrMsg       string `json:"errMsg" gorm:"type:varchar(500);comment:错误"`
}

func (LlmopsAgentEval) TableName() string { return "llmops_agent_eval" }

// LlmopsAgentEvalResult 智能体评测明细
type LlmopsAgentEvalResult struct {
	global.GVA_MODEL
	EvalId   uint    `json:"evalId" gorm:"index;comment:评测任务ID"`
	Idx      int     `json:"idx" gorm:"comment:样本序号"`
	Question string  `json:"question" gorm:"type:varchar(1024);comment:问题"`
	Reply    string  `json:"reply" gorm:"type:text;comment:智能体回答"`
	Score    float64 `json:"score" gorm:"comment:得分0-1"`
	Pass     bool    `json:"pass" gorm:"comment:是否通过"`
	Detail   string  `json:"detail" gorm:"type:varchar(500);comment:评分说明"`
}

func (LlmopsAgentEvalResult) TableName() string { return "llmops_agent_eval_result" }

// LlmopsAgentCallLog 智能体调用日志（E09 调用统计 / E10 运行监控数据源）
type LlmopsAgentCallLog struct {
	global.GVA_MODEL
	AgentId   uint   `json:"agentId" gorm:"index;comment:智能体ID"`
	SessionId uint   `json:"sessionId" gorm:"index;comment:会话ID"`
	LatencyMs int64  `json:"latencyMs" gorm:"comment:耗时毫秒"`
	Ok        bool   `json:"ok" gorm:"index;comment:是否成功"`
	ErrMsg    string `json:"errMsg" gorm:"type:varchar(500);comment:错误"`
}

func (LlmopsAgentCallLog) TableName() string { return "llmops_agent_call_log" }

// LlmopsSkill 技能（要求(6) 技能生态引擎：内置/扩展/自定义）
type LlmopsSkill struct {
	global.GVA_MODEL
	Name    string `json:"name" gorm:"comment:技能名;unique"`
	Type    string `json:"type" gorm:"type:varchar(16);index;comment:builtin/extended/custom"`
	Desc    string `json:"desc" gorm:"type:varchar(500);comment:描述"`
	DefJSON string `json:"defJSON" gorm:"type:text;comment:技能定义JSON"`
	Status  string `json:"status" gorm:"type:varchar(16);default:enabled;index;comment:enabled/disabled"`
}

func (LlmopsSkill) TableName() string { return "llmops_skill" }

// LlmopsMcpTool 业务系统 API 封装的 MCP 工具（要求(6) MCP 协议封装为标准工具）
type LlmopsMcpTool struct {
	global.GVA_MODEL
	Name        string `json:"name" gorm:"comment:工具名;unique"`
	Desc        string `json:"desc" gorm:"type:varchar(500);comment:描述"`
	Endpoint    string `json:"endpoint" gorm:"comment:目标 API 地址"`
	Method      string `json:"method" gorm:"type:varchar(8);default:GET;comment:GET/POST"`
	HeadersJSON string `json:"-" gorm:"type:text;comment:固定头 JSON(值 AES-GCM 密文，不回显)"`
	ArgsSchema  string `json:"argsSchema" gorm:"type:text;comment:参数模板 JSON(字段→说明)"`
	Status      string `json:"status" gorm:"type:varchar(16);default:enabled;index;comment:enabled/disabled"`
}

func (LlmopsMcpTool) TableName() string { return "llmops_mcp_tool" }
