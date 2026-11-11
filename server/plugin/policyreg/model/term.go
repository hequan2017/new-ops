// Package model policyreg 高校制度大模型：制度分词与属性（需求 3.8-4）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// PolicyTermTag 分词与属性标注结果（分类标签/生效日期）
type PolicyTermTag struct {
	global.GVA_MODEL
	DocId         uint   `json:"docId" gorm:"index;comment:制度ID"`
	Tokens        string `json:"tokens" gorm:"type:text;comment:分词结果(JSON 数组)"`
	CategoryLabel string `json:"categoryLabel" gorm:"type:varchar(32);index;comment:分类标签(教学/科研/人事/财务/资产/综合)"`
	EffectiveDate string `json:"effectiveDate" gorm:"type:varchar(32);comment:生效日期(正则抽取)"`
}

func (PolicyTermTag) TableName() string { return "policy_term_tag" }
