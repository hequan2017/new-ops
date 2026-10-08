// Package model aiops AI 运维：LLM 供应商与诊断记录
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// AiopsLlmProvider LLM 供应商（OpenAI 兼容 chat/completions 协议）
type AiopsLlmProvider struct {
	global.GVA_MODEL
	Name       string `json:"name" gorm:"comment:名称;unique"`
	BaseUrl    string `json:"baseUrl" gorm:"comment:API基础地址(含/v1)"`
	ApiKeyEnc  string `json:"-" gorm:"type:text;comment:APIKey密文(AES-GCM)"` // json:"-" 永不序列化
	Model      string `json:"model" gorm:"comment:模型名"`
	IsDefault  bool   `json:"isDefault" gorm:"comment:默认供应商"`
	TimeoutSec int    `json:"timeoutSec" gorm:"comment:调用超时秒数;default:60"`
	Notes      string `json:"notes" gorm:"type:text;comment:备注"`
}

func (AiopsLlmProvider) TableName() string { return "aiops_llm_provider" }

// AiopsDiagnosis Pod 诊断记录（快照+结论留档可审计）
type AiopsDiagnosis struct {
	global.GVA_MODEL
	ClusterID uint   `json:"clusterId" gorm:"comment:集群ID"`
	Cluster   string `json:"cluster" gorm:"comment:集群名"`
	Namespace string `json:"namespace" gorm:"comment:命名空间"`
	Pod       string `json:"pod" gorm:"comment:Pod名"`
	Snapshot  string `json:"snapshot" gorm:"type:longtext;comment:采集快照(状态+事件+日志)"`
	Analysis  string `json:"analysis" gorm:"type:longtext;comment:LLM分析结论"`
	Provider  string `json:"provider" gorm:"comment:供应商名"`
	Model     string `json:"model" gorm:"comment:模型名"`
}

func (AiopsDiagnosis) TableName() string { return "aiops_diagnosis" }
