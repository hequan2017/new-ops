// Package model fanhua 繁花-文科大模型：城市科学知识库（需求 3.4-17 城市治理评估指标树）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// FanhuaGovIndicator 治理评估指标（父子树，含视觉评价维度）
type FanhuaGovIndicator struct {
	global.GVA_MODEL
	Name       string `json:"name" gorm:"type:varchar(100);comment:指标名"`
	ParentId   uint   `json:"parentId" gorm:"default:0;index;comment:父指标ID(0=根)"`
	Weight     int    `json:"weight" gorm:"default:1;comment:同层权重"`
	Visual     bool   `json:"visual" gorm:"comment:是否视觉评价维度"`
	Desc       string `json:"desc" gorm:"type:varchar(300);comment:说明"`
}

func (FanhuaGovIndicator) TableName() string { return "fanhua_gov_indicator" }

// FanhuaGovDataPoint 区域指标数据点（问卷/POI/传感聚合值）
type FanhuaGovDataPoint struct {
	global.GVA_MODEL
	Region      string  `json:"region" gorm:"type:varchar(100);index;comment:区域/社区"`
	IndicatorId uint    `json:"indicatorId" gorm:"index;comment:指标ID"`
	Value       float64 `json:"value" gorm:"comment:聚合值(0-100)"`
	Period      string  `json:"period" gorm:"type:varchar(7);index;comment:期间 YYYY-MM"`
	Source      string  `json:"source" gorm:"type:varchar(16);default:questionnaire;comment:questionnaire/poi/sensor"`
}

func (FanhuaGovDataPoint) TableName() string { return "fanhua_gov_data_point" }
