// Package model fanhua 繁花-文科大模型：图片标注（需求 3.4-13/14 图片训练/测试数据标注）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// FanhuaImageAsset 图片资产登记（train/test 用途字段）
type FanhuaImageAsset struct {
	global.GVA_MODEL
	Name     string `json:"name" gorm:"type:varchar(200);comment:图片名称"`
	Path     string `json:"path" gorm:"type:varchar(500);comment:存储路径/URL"`
	MIME     string `json:"mime" gorm:"type:varchar(64);comment:MIME 类型"`
	Source   string `json:"source" gorm:"type:varchar(16);default:upload;comment:upload/url"`
	Split    string `json:"split" gorm:"type:varchar(8);default:none;index;comment:train/test/none 用途"`
	MetaJSON string `json:"metaJSON" gorm:"type:text;comment:元数据"`
}

func (FanhuaImageAsset) TableName() string { return "fanhua_image_asset" }

// FanhuaImageLabel 图片标签记录（一图一条 upsert）
type FanhuaImageLabel struct {
	global.GVA_MODEL
	AssetId uint   `json:"assetId" gorm:"uniqueIndex;comment:图片资产ID"`
	Tags    string `json:"tags" gorm:"type:varchar(500);comment:标签(JSON 数组)"`
	Labeled bool   `json:"labeled" gorm:"index;comment:是否已标注"`
}

func (FanhuaImageLabel) TableName() string { return "fanhua_image_label" }
