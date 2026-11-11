// Package model crossauth 跨校访问认证与安全：数据分级引擎（需求 3.7-9）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CaDataClassify 分级结果（分类×分级）
type CaDataClassify struct {
	global.GVA_MODEL
	ResourceName string `json:"resourceName" gorm:"type:varchar(300);index;comment:资源名/内容摘要"`
	Category     string `json:"category" gorm:"type:varchar(16);index;comment:教学/科研/管理"`
	Level        string `json:"level" gorm:"type:varchar(16);index;comment:公开/内部/敏感"`
	HitRule      string `json:"hitRule" gorm:"type:varchar(200);comment:命中规则"`
	ScannedAt    string `json:"scannedAt" gorm:"type:varchar(32);comment:扫描时间"`
}

func (CaDataClassify) TableName() string { return "ca_data_classify" }
