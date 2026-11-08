// Package model fanhua 繁花-文科大模型：双语平行语料（需求 3.4-4 对照本收集 / 3.4-15 三语平行语料库）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// FanhuaParallel 沪/普/英三语平行语料
type FanhuaParallel struct {
	global.GVA_MODEL
	Sh     string `json:"sh" gorm:"type:varchar(500);comment:沪语句"`
	Pu     string `json:"pu" gorm:"type:varchar(500);index;comment:普通话句"`
	En     string `json:"en" gorm:"type:varchar(500);comment:英语句"`
	Source string `json:"source" gorm:"type:varchar(100);comment:来源(对照本/教学材料等)"`
}

func (FanhuaParallel) TableName() string { return "fanhua_parallel" }
