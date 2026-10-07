// Package model 白泽批量作业数据模型（M2：脚本库与变量组）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// 脚本语言
const (
	ScriptLangShell = "shell"
	ScriptLangPython = "python"
	ScriptLangYaml   = "yaml"
)

// ValidScriptLangs 合法脚本语言集合
func ValidScriptLangs() map[string]bool {
	return map[string]bool{ScriptLangShell: true, ScriptLangPython: true, ScriptLangYaml: true}
}

// JobScript 脚本库条目（content 每次更新递增版本并落版本表）
type JobScript struct {
	global.GVA_MODEL
	Name     string `json:"name" gorm:"comment:脚本名称;unique" binding:"required"`
	Language string `json:"language" gorm:"comment:语言(shell/python/yaml)" binding:"required"`
	Content  string `json:"content" gorm:"type:longtext;comment:当前版本内容"`
	Version  int    `json:"version" gorm:"comment:当前版本号;default:1"`
	Notes    string `json:"notes" gorm:"type:text;comment:备注"`
}

// JobScriptVersion 脚本历史版本（更新时归档旧内容）
type JobScriptVersion struct {
	global.GVA_MODEL
	ScriptID uint   `json:"scriptId" gorm:"comment:脚本ID;index"`
	Version  int    `json:"version" gorm:"comment:版本号"`
	Content  string `json:"content" gorm:"type:longtext;comment:该版本内容"`
	Operator string `json:"operator" gorm:"comment:操作人"`
}

// JobVariableGroup 变量组（关联资产组，执行时按主机所在组注入变量）
type JobVariableGroup struct {
	global.GVA_MODEL
	Name         string `json:"name" gorm:"comment:变量组名称;unique" binding:"required"`
	Variables    string `json:"variables" gorm:"type:text;comment:变量JSON([{key,value}])"`
	AssetGroupID *uint  `json:"assetGroupId" gorm:"comment:关联资产组ID(空则全局)"`
	Notes        string `json:"notes" gorm:"type:text;comment:备注"`
}
