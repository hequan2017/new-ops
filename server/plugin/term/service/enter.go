// Package service term 插件服务组
package service

// ServiceGroup term 插件服务组
type ServiceGroup struct {
	TermAuditService
	SftpService
}

var Service = new(ServiceGroup)
