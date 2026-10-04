// Package model 白泽流水线数据模型（M3）
// 三层模型（new-jenkins 蓝本）：Pipeline 1-N Stage 1-N Step；
// 触发快照与构建状态机随后续场次（pipeline_build/build_log）。
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// 步骤类型
const (
	StepTypeShell = "shell"
	StepTypeHTTP  = "http"
)

// ValidStepTypes 合法步骤类型集合
func ValidStepTypes() map[string]bool {
	return map[string]bool{StepTypeShell: true, StepTypeHTTP: true}
}

// Pipeline 流水线定义
type Pipeline struct {
	global.GVA_MODEL
	Name           string          `json:"name" gorm:"comment:流水线名称;unique" binding:"required"`
	Description    string          `json:"description" gorm:"comment:描述"`
	Enabled        bool            `json:"enabled" gorm:"comment:是否启用;default:true"`
	WebhookEnabled bool            `json:"webhookEnabled" gorm:"comment:启用webhook触发"`
	WebhookToken   string          `json:"webhookToken" gorm:"size:64;comment:webhook令牌(即凭据,仅开启时生成)"`
	CronEnabled    bool            `json:"cronEnabled" gorm:"comment:启用定时触发"`
	CronSpec       string          `json:"cronSpec" gorm:"size:32;comment:cron表达式(robfig标准5段)"`
	Stages         []PipelineStage `json:"stages" gorm:"foreignKey:PipelineID"`
}

// PipelineStage 阶段（串行执行，Sort 升序）
type PipelineStage struct {
	global.GVA_MODEL
	PipelineID      uint   `json:"pipelineId" gorm:"comment:流水线ID;index"`
	Name            string `json:"name" gorm:"comment:阶段名称" binding:"required"`
	Sort            int    `json:"sort" gorm:"comment:顺序"`
	Approval        bool   `json:"approval" gorm:"comment:人工审批gate"`
	ContinueOnError bool   `json:"continueOnError" gorm:"comment:失败继续"`
	Steps           []PipelineStep `json:"steps" gorm:"foreignKey:StageID"`
}

// PipelineStep 步骤（阶段内串行）
type PipelineStep struct {
	global.GVA_MODEL
	StageID      uint   `json:"stageId" gorm:"comment:阶段ID;index"`
	Name         string `json:"name" gorm:"comment:步骤名称" binding:"required"`
	Sort         int    `json:"sort" gorm:"comment:顺序"`
	Type         string `json:"type" gorm:"comment:类型(shell/http)"`
	HostID       uint   `json:"hostId" gorm:"comment:目标主机ID(shell步骤在资产上SSH执行)"`
	ShellContent string `json:"shellContent" gorm:"type:text;comment:shell内容"`
	HTTPMethod   string `json:"httpMethod" gorm:"comment:HTTP方法"`
	HTTPURL      string `json:"httpUrl" gorm:"comment:HTTP地址"`
	HTTPBody     string `json:"httpBody" gorm:"type:text;comment:HTTP体"`
	TimeoutSec   int    `json:"timeoutSec" gorm:"comment:超时秒(0不限)"`
}
