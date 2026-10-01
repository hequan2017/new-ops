package service

// ServiceGroup asset 插件服务组
type ServiceGroup struct {
	AssetHostService
	AssetRoomService
	AssetRackService
	AssetProductLineService
}

var Service = new(ServiceGroup)
