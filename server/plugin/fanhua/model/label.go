// Package model fanhua 繁花-文科大模型：文本标注（需求 3.4-11/12 训练/测试数据标注）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// FanhuaTextLabel 文本标注记录
type FanhuaTextLabel struct {
	global.GVA_MODEL
	CorpusId      uint   `json:"corpusId" gorm:"default:0;index;comment:关联语料ID(0=自由文本)"`
	Text          string `json:"text" gorm:"type:text;comment:标注文本"`
	Split         string `json:"split" gorm:"type:varchar(8);index;comment:train/test/none"`
	AutoSentiment string `json:"autoSentiment" gorm:"type:varchar(8);comment:规则预标(正面/中立/负面)"`
	Sentiment     string `json:"sentiment" gorm:"type:varchar(8);index;comment:最终标签(人工改标后)"`
	Labeled       bool   `json:"labeled" gorm:"index;comment:是否人工改标"`
}

func (FanhuaTextLabel) TableName() string { return "fanhua_text_label" }
