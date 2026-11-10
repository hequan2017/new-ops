// Package model ziqiang 自强-理工科大模型：侵权检测提示词库（需求 3.5-12/13）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ZqPromptTemplate 提示词模板（{{var}} 占位，对齐 llmops G01/N12 口径）
// scene=多模态双验证（3.5-12）/引导式对话（3.5-13，step 1-3：创作场景→权利类型→法律文书）
type ZqPromptTemplate struct {
	global.GVA_MODEL
	Name    string `json:"name" gorm:"type:varchar(200);uniqueIndex;comment:模板名"`
	Scene   string `json:"scene" gorm:"type:varchar(32);index;comment:多模态双验证/引导式对话"`
	Step    int    `json:"step" gorm:"default:1;comment:引导式对话步号(1-3),单模板为 1"`
	Content string `json:"content" gorm:"type:text;comment:模板内容({{var}})"`
	Vars    string `json:"vars" gorm:"type:varchar(500);comment:占位变量(JSON 数组)"`
	Enabled bool   `json:"enabled" gorm:"default:true;index;comment:启用"`
}

func (ZqPromptTemplate) TableName() string { return "ziqiang_prompt_template" }
