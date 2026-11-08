// Package model fanhua 繁花-文科大模型：价值观审查日志（需求 3.4-21）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// FanhuaValuesAuditLog 价值观合规审查日志（中英输入，时延 <800ms 口径）
type FanhuaValuesAuditLog struct {
	global.GVA_MODEL
	Text          string `json:"text" gorm:"type:mediumtext;comment:审查文本"`
	RiskLevel     string `json:"riskLevel" gorm:"type:varchar(8);index;comment:高/中/低"`
	Health        int    `json:"health" gorm:"comment:健康度评分"`
	Sentiment     string `json:"sentiment" gorm:"type:varchar(8);comment:情感三分类"`
	Blocked       bool   `json:"blocked" gorm:"comment:是否拦截"`
	Alarm         bool   `json:"alarm" gorm:"index;comment:异常预警信号"`
	HitsJSON      string `json:"hitsJSON" gorm:"type:text;comment:命中明细(JSON)"`
	LatencyMs     int64  `json:"latencyMs" gorm:"comment:语义解析时延(毫秒)"`
	OverThreshold bool   `json:"overThreshold" gorm:"comment:超 800ms 阈值"`
}

func (FanhuaValuesAuditLog) TableName() string { return "fanhua_values_audit_log" }
