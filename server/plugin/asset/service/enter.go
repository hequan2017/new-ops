package service

// ServiceGroup asset 插件服务组
type ServiceGroup struct {
	AssetHostService
}

var Service = new(ServiceGroup)
