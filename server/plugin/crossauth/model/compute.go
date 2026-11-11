// Package model crossauth 跨校访问认证与安全：算力使用防护（需求 3.7-12）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CaComputeUsage 算力计量（单用户×小时窗口，单用户 1 小时 1000 核时口径）
type CaComputeUsage struct {
	global.GVA_MODEL
	UserName    string  `json:"userName" gorm:"type:varchar(100);uniqueIndex:idx_ca_usage_user_hour;comment:用户"`
	WindowHour  string  `json:"windowHour" gorm:"type:varchar(16);uniqueIndex:idx_ca_usage_user_hour;comment:小时窗口(yyyy-MM-dd HH)"`
	GpuHours    float64 `json:"gpuHours" gorm:"comment:GPU 卡时累计"`
	CoreHours   float64 `json:"coreHours" gorm:"comment:核时累计"`
}

func (CaComputeUsage) TableName() string { return "ca_compute_usage" }

// CaComputeAlert 算力告警（限流/降优先级）
type CaComputeAlert struct {
	global.GVA_MODEL
	UserName string  `json:"userName" gorm:"type:varchar(100);index;comment:用户"`
	WindowHour string `json:"windowHour" gorm:"type:varchar(16);comment:小时窗口"`
	CoreHours float64 `json:"coreHours" gorm:"comment:核时"`
	Action   string  `json:"action" gorm:"type:varchar(16);comment:限流/降优先级"`
	Reason   string  `json:"reason" gorm:"type:varchar(300);comment:说明"`
	Handled  bool    `json:"handled" gorm:"index;comment:管理员已处置"`
}

func (CaComputeAlert) TableName() string { return "ca_compute_alert" }
