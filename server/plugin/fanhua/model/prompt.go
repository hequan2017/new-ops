// Package model fanhua 繁花-文科大模型：提示词库（需求 3.4-18/19）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// FanhuaPromptTemplate 提示词模板（{{var}} 占位，对齐 llmops G01 口径）
type FanhuaPromptTemplate struct {
	global.GVA_MODEL
	Name    string `json:"name" gorm:"type:varchar(200);uniqueIndex;comment:模板名"`
	Scene   string `json:"scene" gorm:"type:varchar(32);index;comment:沪语场景/价值观审核/通用"`
	Content string `json:"content" gorm:"type:text;comment:模板内容({{var}})"`
	Vars    string `json:"vars" gorm:"type:varchar(500);comment:占位变量(JSON 数组)"`
	Enabled bool   `json:"enabled" gorm:"default:true;index;comment:启用"`
}

func (FanhuaPromptTemplate) TableName() string { return "fanhua_prompt_template" }
