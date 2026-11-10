// Package model ziqiang 自强-理工科大模型：文本鉴权语料（需求 3.5-1）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ZqTextCorpus 知识产权文本语料（专利/开源协议/论文/合同，AIGC 对比样本）
type ZqTextCorpus struct {
	global.GVA_MODEL
	Type          string `json:"type" gorm:"type:varchar(16);index;comment:专利/开源协议/论文/合同"`
	Title         string `json:"title" gorm:"type:varchar(300);comment:标题"`
	Content       string `json:"content" gorm:"type:mediumtext;comment:正文"`
	AigcFlag      bool   `json:"aigcFlag" gorm:"index;comment:是否 AIGC 生成"`
	Source        string `json:"source" gorm:"type:varchar(100);comment:来源(专利号/平台/期刊)"`
	LegalStatus   string `json:"legalStatus" gorm:"type:varchar(16);default:有效;index;comment:法律状态(有效/审查中/失效/届满)"`
	ImageCorpusId uint   `json:"imageCorpusId" gorm:"default:0;comment:关联图像语料ID(融合检索图像维度)"`
}

func (ZqTextCorpus) TableName() string { return "ziqiang_text_corpus" }

// ZqComparePair AIGC 对比样本对（人创/AIGC 成对）
type ZqComparePair struct {
	global.GVA_MODEL
	HumanCorpusId uint   `json:"humanCorpusId" gorm:"index;comment:人创语料ID"`
	AigcCorpusId  uint   `json:"aigcCorpusId" gorm:"index;comment:AIGC 语料ID"`
	Domain        string `json:"domain" gorm:"type:varchar(50);comment:领域"`
	Note          string `json:"note" gorm:"type:varchar(300);comment:说明"`
}

func (ZqComparePair) TableName() string { return "ziqiang_compare_pair" }
