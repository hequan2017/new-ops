// Package model ziqiang 自强-理工科大模型：知识产权实体图谱（需求 3.5-6）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ZqEntity 知识产权实体（法律[专利号,商标注册码]/技术[模型架构,训练方法]/权利主体[开发者,投资方]）
type ZqEntity struct {
	global.GVA_MODEL
	Name     string `json:"name" gorm:"type:varchar(200);uniqueIndex;comment:实体名"`
	Category string `json:"category" gorm:"type:varchar(16);index;comment:法律/技术/权利主体"`
	Kind     string `json:"kind" gorm:"type:varchar(32);comment:专利号/商标注册码/模型架构/训练方法/开发者/投资方"`
	Desc     string `json:"desc" gorm:"type:varchar(300);comment:说明"`
}

func (ZqEntity) TableName() string { return "ziqiang_entity" }

// ZqEdge 实体关系边
type ZqEdge struct {
	global.GVA_MODEL
	FromId   uint   `json:"fromId" gorm:"index;comment:起点实体ID"`
	ToId     uint   `json:"toId" gorm:"index;comment:终点实体ID"`
	Relation string `json:"relation" gorm:"type:varchar(32);comment:申请/持有/引用/投资/许可/涉及技术"`
}

func (ZqEdge) TableName() string { return "ziqiang_edge" }
