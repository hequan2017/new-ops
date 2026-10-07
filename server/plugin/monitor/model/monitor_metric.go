// Package model 白泽监控指标数据模型（M6：SSH 性能采集）
// 指标名约定：cpu_percent / mem_percent / disk_percent / load1（DEV_PLAN 3.4）
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// MetricName 指标名集合
const (
	MetricCPUPercent  = "cpu_percent"
	MetricMemPercent  = "mem_percent"
	MetricDiskPercent = "disk_percent"
	MetricLoad1       = "load1"
	MetricNetRxKBs    = "net_rx_kbs"
	MetricNetTxKBs    = "net_tx_kbs"
)

// ValidMetricNames 合法指标集合（查询过滤白名单）
func ValidMetricNames() map[string]bool {
	return map[string]bool{
		MetricCPUPercent: true, MetricMemPercent: true,
		MetricDiskPercent: true, MetricLoad1: true,
		MetricNetRxKBs: true, MetricNetTxKBs: true,
	}
}

// MonitorMetric 性能指标点
type MonitorMetric struct {
	global.GVA_MODEL
	AssetID uint      `json:"assetId" gorm:"comment:资产ID;index:idx_metric_asset_name_ts,priority:1"`
	Name    string    `json:"name" gorm:"size:32;comment:指标名;index:idx_metric_asset_name_ts,priority:2"`
	Value   float64   `json:"value" gorm:"comment:指标值"`
	TS      time.Time `json:"ts" gorm:"comment:采集时间;index:idx_metric_asset_name_ts,priority:3"`
}
