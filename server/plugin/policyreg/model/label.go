// Package model policyreg 高校制度大模型：训练数据标注（需求 3.8-6）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// PolicyTokenLabel 分词实体标注（规则预标/人工改标）
type PolicyTokenLabel struct {
	global.GVA_MODEL
	DocId  uint   `json:"docId" gorm:"index;comment:制度ID"`
	Token  string `json:"token" gorm:"type:varchar(200);comment:词项"`
	Label  string `json:"label" gorm:"type:varchar(32);index;comment:O/机构/文件号/日期/专有名词"`
	Source string `json:"source" gorm:"type:varchar(16);comment:规则预标/人工"`
	Seq    int    `json:"seq" gorm:"comment:词序"`
}

func (PolicyTokenLabel) TableName() string { return "policy_token_label" }

// PolicyRelationLabel 关系标注（机构×文件号 共现规则）
type PolicyRelationLabel struct {
	global.GVA_MODEL
	DocId    uint   `json:"docId" gorm:"index;comment:制度ID"`
	Head     string `json:"head" gorm:"type:varchar(200);comment:头实体(机构)"`
	Tail     string `json:"tail" gorm:"type:varchar(200);comment:尾实体(文件号)"`
	Relation string `json:"relation" gorm:"type:varchar(32);comment:发文机构/涉及机构"`
	Source   string `json:"source" gorm:"type:varchar(16);comment:规则预标/人工"`
}

func (PolicyRelationLabel) TableName() string { return "policy_relation_label" }
