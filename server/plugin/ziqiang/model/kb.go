// Package model ziqiang 自强-理工科大模型：模型鉴权知识库（需求 3.5-7）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ZqRegulation 法规条目（中美欧三地）
type ZqRegulation struct {
	global.GVA_MODEL
	Jurisdiction  string `json:"jurisdiction" gorm:"type:varchar(16);index;comment:中国/美国/欧盟"`
	Category      string `json:"category" gorm:"type:varchar(16);index;comment:版权/专利/商标/数据合规/AI治理"`
	Title         string `json:"title" gorm:"type:varchar(300);comment:法规名"`
	Content       string `json:"content" gorm:"type:mediumtext;comment:要点内容"`
	EffectiveDate string `json:"effectiveDate" gorm:"type:varchar(16);comment:生效日期"`
}

func (ZqRegulation) TableName() string { return "ziqiang_regulation" }

// ZqCase 判例条目
type ZqCase struct {
	global.GVA_MODEL
	Jurisdiction string `json:"jurisdiction" gorm:"type:varchar(16);index;comment:中国/美国/欧盟"`
	Name         string `json:"name" gorm:"type:varchar(300);comment:案件名"`
	Summary      string `json:"summary" gorm:"type:text;comment:案情概要"`
	Ruling       string `json:"ruling" gorm:"type:text;comment:裁判要点"`
	Year         int    `json:"year" gorm:"comment:裁判年份"`
}

func (ZqCase) TableName() string { return "ziqiang_case" }
