// Package model eduadmin 教育管理一模通管：智能采购（需求 3.6-19）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduPurchase 历史采购记录（按月）
type EduPurchase struct {
	global.GVA_MODEL
	Year      int     `json:"year" gorm:"index;comment:年度"`
	Month     int     `json:"month" gorm:"index;comment:月份(1-12)"`
	Item      string  `json:"item" gorm:"type:varchar(200);index;comment:采购品目"`
	Qty       float64 `json:"qty" gorm:"comment:采购数量"`
	UnitPrice float64 `json:"unitPrice" gorm:"comment:单价(元)"`
}

func (EduPurchase) TableName() string { return "edu_purchase" }
