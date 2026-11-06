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
