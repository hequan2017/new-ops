// Package model eduadmin 教育管理一模通管：食堂优化（需求 3.6-21）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduMealRecord 一卡通消费记录（档口/餐段）
type EduMealRecord struct {
	global.GVA_MODEL
	Date     string  `json:"date" gorm:"type:varchar(16);index;comment:日期(yyyy-mm-dd)"`
	Stall    string  `json:"stall" gorm:"type:varchar(100);index;comment:档口"`
	MealTime string  `json:"mealTime" gorm:"type:varchar(16);comment:早餐/午餐/晚餐"`
	Count    int     `json:"count" gorm:"comment:用餐人次"`
	Amount   float64 `json:"amount" gorm:"comment:消费金额(元)"`
}

func (EduMealRecord) TableName() string { return "edu_meal_record" }
