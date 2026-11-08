// Package model artsci 学科工具集：图像解读与分析（需求 3.3-13）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ArtsciImageAsset 图像资产登记（元数据口径）
type ArtsciImageAsset struct {
	global.GVA_MODEL
	Name     string `json:"name" gorm:"type:varchar(200);comment:图像名称"`
	Path     string `json:"path" gorm:"type:varchar(500);comment:存储路径/URL"`
	MIME     string `json:"mime" gorm:"type:varchar(64);comment:MIME 类型"`
	SizeByte int64  `json:"sizeByte" gorm:"comment:大小(字节)"`
	Source   string `json:"source" gorm:"type:varchar(16);default:upload;comment:upload/url/generated"`
	MetaJSON string `json:"metaJSON" gorm:"type:text;comment:来源元数据(创作者/年代等)"`
}

func (ArtsciImageAsset) TableName() string { return "artsci_image_asset" }

// ArtsciInterpretation 图像解读记录（结构化：表面细节 + 文化含义）
type ArtsciInterpretation struct {
	global.GVA_MODEL
	AssetId         uint   `json:"assetId" gorm:"index;comment:图像资产ID"`
	Subject         string `json:"subject" gorm:"type:varchar(200);comment:解读视角(艺术作品/历史照片等)"`
	SurfaceDesc     string `json:"surfaceDesc" gorm:"type:text;comment:表面细节描述"`
	CulturalMeaning string `json:"culturalMeaning" gorm:"type:text;comment:文化/社会/历史含义"`
	Engine          string `json:"engine" gorm:"type:varchar(16);default:rule;comment:llm/rule"`
	ModelId         uint   `json:"modelId" gorm:"default:0;comment:llmops 模型ID(0=元数据口径)"`
	Ok              bool   `json:"ok" gorm:"index;comment:解读是否成功"`
	ErrMsg          string `json:"errMsg" gorm:"type:varchar(500);comment:错误"`
}

func (ArtsciInterpretation) TableName() string { return "artsci_interpretation" }
