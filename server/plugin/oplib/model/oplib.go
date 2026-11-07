// Package model oplib 算子库：算子注册
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// OplibOperator 算子（采购件② 算子库；H03 执行/H04 编排复用）
type OplibOperator struct {
	global.GVA_MODEL
	Name         string `json:"name" gorm:"comment:算子名;unique"`
	Category     string `json:"category" gorm:"type:varchar(16);index;comment:数据处理/特征/检索/后处理/评估"`
	ImplType     string `json:"implType" gorm:"type:varchar(16);default:builtin;index;comment:builtin/script"`
	InputSchema  string `json:"inputSchema" gorm:"type:text;comment:入参schema JSON(字段→类型)"`
	OutputSchema string `json:"outputSchema" gorm:"type:text;comment:出参schema JSON"`
	ScriptLang   string `json:"scriptLang" gorm:"type:varchar(16);comment:脚本解释器(python3/node，script 型)"`
	ScriptCode   string `json:"scriptCode" gorm:"type:mediumtext;comment:脚本代码(script 型)"`
	Version      string `json:"version" gorm:"type:varchar(16);default:v1;comment:版本"`
	Desc         string `json:"desc" gorm:"type:varchar(500);comment:描述"`
}

func (OplibOperator) TableName() string { return "oplib_operator" }
