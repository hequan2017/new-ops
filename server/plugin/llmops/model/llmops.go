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
