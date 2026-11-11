// Package model policyreg 高校制度大模型：提示词库与问答（需求 3.8-8）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// PolicyPromptTemplate 制度问答提示词模板（{{var}} 占位）
type PolicyPromptTemplate struct {
	global.GVA_MODEL
	Name    string `json:"name" gorm:"type:varchar(200);uniqueIndex;comment:模板名"`
	Content string `json:"content" gorm:"type:text;comment:模板内容({{var}})"`
	Vars    string `json:"vars" gorm:"type:varchar(500);comment:占位变量(JSON 数组)"`
	Enabled bool   `json:"enabled" gorm:"default:true;index;comment:启用"`
}

func (PolicyPromptTemplate) TableName() string { return "policy_prompt_template" }

// PolicyQaLog 制度问答日志
type PolicyQaLog struct {
	global.GVA_MODEL
	Question  string `json:"question" gorm:"type:varchar(500);comment:问题"`
	Answer    string `json:"answer" gorm:"type:text;comment:回答"`
	Source    string `json:"source" gorm:"type:varchar(16);comment:chat/规则"`
	LatencyMs int64  `json:"latencyMs" gorm:"comment:时延(ms)"`
	RoleName  string `json:"roleName" gorm:"type:varchar(100);comment:提问角色(鉴权过滤口径)"`
}

func (PolicyQaLog) TableName() string { return "policy_qa_log" }
