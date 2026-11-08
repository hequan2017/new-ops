// Package model fanhua 繁花-文科大模型：城市科学融合查询（需求 3.4-22）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// FanhuaGovPhoto 社区照片元数据（第三源）
type FanhuaGovPhoto struct {
	global.GVA_MODEL
	Region  string `json:"region" gorm:"type:varchar(100);index;comment:区域"`
	Path    string `json:"path" gorm:"type:varchar(500);comment:照片路径/URL"`
	ShotAt  string `json:"shotAt" gorm:"type:varchar(19);comment:拍摄时间"`
	MetaJSON string `json:"metaJSON" gorm:"type:text;comment:拍摄元数据"`
}

func (FanhuaGovPhoto) TableName() string { return "fanhua_gov_photo" }
