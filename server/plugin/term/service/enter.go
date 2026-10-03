package service

// ServiceGroup term 插件服务组
type ServiceGroup struct {
	TermAuditService
}

var Service = new(ServiceGroup)
