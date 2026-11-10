// Package model eduadmin 教育管理一模通管：能耗优化（需求 3.6-20）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduEnergyRecord 能耗记录（楼栋/年月/时段）
type EduEnergyRecord struct {
	global.GVA_MODEL
	Building string  `json:"building" gorm:"type:varchar(100);index;comment:楼栋"`
	Year     int     `json:"year" gorm:"index;comment:年度"`
	Month    int     `json:"month" gorm:"index;comment:月份(1-12)"`
	Period   string  `json:"period" gorm:"type:varchar(16);comment:白天/夜间"`
	Kwh      float64 `json:"kwh" gorm:"comment:耗电量(度)"`
}

func (EduEnergyRecord) TableName() string { return "edu_energy_record" }
