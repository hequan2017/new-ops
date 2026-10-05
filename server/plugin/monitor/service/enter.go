// Package service monitor 插件服务出口与错误码
package service

// MonitorService 监控服务
type MonitorService struct{}

// Service 聚合出口
var Service = new(containerlessSvc)

type containerlessSvc struct {
	Monitor MonitorService
}

// monitor 插件错误码段：1800-1899（DEV_PLAN 3.6）
const (
	ErrCodeMetricNameInvalid = 1801
)
