// Package model 白泽监控告警（M6：规则引擎/事件）
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// 规则类型
const (
	RuleTypeMetric = "metric" // 指标阈值（cpu_percent 等）
	RuleTypePort   = "port"   // 端口探活（TCP 可达）
)

// 比较符
const (
	OpGT = ">"
	OpLT = "<"
)

// MonitorAlertRule 告警规则
type MonitorAlertRule struct {
	global.GVA_MODEL
	AssetID    uint   `json:"assetId" gorm:"comment:资产ID;index"`
	Type       string `json:"type" gorm:"comment:类型(metric/port)"`
	MetricName string `json:"metricName" gorm:"size:32;comment:指标名(metric类型)"`
	Operator   string `json:"operator" gorm:"size:2;comment:比较符(>/ <)"`
	Threshold  float64 `json:"threshold" gorm:"comment:阈值"`
	Port       int    `json:"port" gorm:"comment:端口(port类型)"`
	Duration   int    `json:"duration" gorm:"comment:持续采样次数(默认1)"`
	SilenceMin int    `json:"silenceMin" gorm:"comment:静默窗口分钟(默认30)"`
	WebhookURL string `json:"webhookUrl" gorm:"size:255;comment:钉钉机器人webhook(空则仅记录事件)"`
	Enabled    bool   `json:"enabled" gorm:"comment:启用;default:true"`
	Notes      string `json:"notes" gorm:"comment:备注"`
}

// MonitorAlertEvent 告警事件（触发与恢复）
type MonitorAlertEvent struct {
	global.GVA_MODEL
	RuleID    uint       `json:"ruleId" gorm:"comment:规则ID;index"`
	AssetID   uint       `json:"assetId" gorm:"comment:资产ID;index"`
	Summary   string     `json:"summary" gorm:"comment:事件摘要"`
	Value     float64    `json:"value" gorm:"comment:触发值"`
	FiredAt   time.Time  `json:"firedAt" gorm:"comment:触发时间"`
	ResolvedAt *time.Time `json:"resolvedAt" gorm:"comment:恢复时间"`
	Notified  bool       `json:"notified" gorm:"comment:已推送通知"`
}
