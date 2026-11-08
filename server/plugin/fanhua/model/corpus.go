// Package model fanhua 繁花-文科大模型：沪语语料库（需求 3.4-1~6 数据收集）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// FanhuaCorpus 沪语语料（六源：新闻/社交/文学/对照/舆情/城市）
type FanhuaCorpus struct {
	global.GVA_MODEL
	Source   string `json:"source" gorm:"type:varchar(16);index;comment:新闻/社交/文学/对照/舆情/城市"`
	Title    string `json:"title" gorm:"type:varchar(300);comment:标题"`
	Content  string `json:"content" gorm:"type:mediumtext;comment:正文"`
	Lang     string `json:"lang" gorm:"type:varchar(16);default:沪语;comment:沪语/普通话/英语/混合"`
	MetaJSON string `json:"metaJSON" gorm:"type:text;comment:来源元数据(渠道/时间等)"`
}

func (FanhuaCorpus) TableName() string { return "fanhua_corpus" }
