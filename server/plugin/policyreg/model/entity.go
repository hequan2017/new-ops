// Package model policyreg 高校制度大模型：实体构建（需求 3.8-5）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// PolicyEntity 制度实体（机构/文件号/日期/专有名词）
type PolicyEntity struct {
	global.GVA_MODEL
	DocId  uint   `json:"docId" gorm:"index;comment:制度ID"`
	Name   string `json:"name" gorm:"type:varchar(200);comment:实体名"`
	EType  string `json:"eType" gorm:"type:varchar(32);index;comment:机构/文件号/日期/专有名词"`
	Source string `json:"source" gorm:"type:varchar(16);comment:词典/规则"`
}

func (PolicyEntity) TableName() string { return "policy_entity" }
