// Package model artsci 学科工具集：工具注册
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ArtsciTool 学科工具（M02 注册表；M03+ 内置实现按 module 分发）
type ArtsciTool struct {
	global.GVA_MODEL
	Name         string `json:"name" gorm:"comment:工具名;unique"`
	Subject      string `json:"subject" gorm:"type:varchar(16);index;comment:数学/文科/图像"`
	Module       string `json:"module" gorm:"type:varchar(32);index;comment:内置实现模块标识"`
	ParamsSchema string `json:"paramsSchema" gorm:"type:text;comment:参数schema JSON(type/required/enum)"`
	Status       string `json:"status" gorm:"type:varchar(16);default:enabled;index;comment:enabled/disabled"`
	Desc         string `json:"desc" gorm:"type:varchar(500);comment:描述"`
}

func (ArtsciTool) TableName() string { return "artsci_tool" }
