// Package model 白泽流水线构建数据模型（M3 场26：执行器）
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// 构建状态机：等待中 → 执行中 →（等待审批）→ 成功 | 失败 | 已取消
const (
	BuildPending   = "等待中"
	BuildAwaiting  = "等待审批"
	BuildRunning   = "执行中"
	BuildSuccess   = "成功"
	BuildFailed    = "失败"
	BuildCanceled  = "已取消"
)

// 构建日志级别
const (
	LogSystem = "system"
	LogStdout = "stdout"
	LogStderr = "stderr"
)

// PipelineBuild 构建记录（触发时对定义做快照，之后改定义不影响本次构建）
type PipelineBuild struct {
	global.GVA_MODEL
	PipelineID   uint   `json:"pipelineId" gorm:"comment:流水线ID;index"`
	PipelineName string `json:"pipelineName" gorm:"comment:流水线名快照"`
	BuildNo      int    `json:"buildNo" gorm:"comment:构建序号(按流水线递增)"`
	Status       string `json:"status" gorm:"comment:状态;index"`
	Params       string `json:"params" gorm:"type:text;comment:触发参数JSON"`
	Snapshot     string `json:"snapshot" gorm:"type:longtext;comment:定义快照JSON(变量已替换)"`
	Operator     string `json:"operator" gorm:"comment:触发人"`
	UserID       uint   `json:"userId" gorm:"comment:触发用户ID;index"`
	StartedAt    *time.Time `json:"startedAt" gorm:"comment:开始时间"`
	FinishedAt   *time.Time `json:"finishedAt" gorm:"comment:结束时间"`
}

// BuildLog 构建日志（按阶段/步骤归类）
type BuildLog struct {
	global.GVA_MODEL
	BuildID   uint   `json:"buildId" gorm:"comment:构建ID;index"`
	StageName string `json:"stageName" gorm:"comment:阶段名"`
	StepName  string `json:"stepName" gorm:"comment:步骤名"`
	Level     string `json:"level" gorm:"comment:级别(system/stdout/stderr)"`
	Content   string `json:"content" gorm:"type:longtext;comment:内容"`
}
