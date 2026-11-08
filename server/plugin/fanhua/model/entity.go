// Package model fanhua 繁花-文科大模型：海派文化实体库（需求 3.4-10 实体构建）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// FanhuaEntity 海派文化实体（地域/文化符号/地标，词条+别名）
type FanhuaEntity struct {
	global.GVA_MODEL
	Category  string `json:"category" gorm:"type:varchar(16);index;comment:地域/文化符号/地标"`
	Term      string `json:"term" gorm:"type:varchar(100);uniqueIndex;comment:词条"`
	AliasJSON string `json:"aliasJSON" gorm:"type:text;comment:别名(JSON 数组)"`
	Desc      string `json:"desc" gorm:"type:varchar(500);comment:释义"`
}

func (FanhuaEntity) TableName() string { return "fanhua_entity" }

// FanhuaEntityHit 词典匹配标注结果
type FanhuaEntityHit struct {
	global.GVA_MODEL
	CorpusId uint   `json:"corpusId" gorm:"default:0;index;comment:关联语料ID(0=直接文本)"`
	Text     string `json:"text" gorm:"type:mediumtext;comment:标注文本"`
	HitsJSON string `json:"hitsJSON" gorm:"type:text;comment:命中结果(JSON)"`
}

func (FanhuaEntityHit) TableName() string { return "fanhua_entity_hit" }
