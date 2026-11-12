// Package service 白泽公告服务聚合
package service

// Service 服务聚合（InfoService 零值可用：db 为 nil 惰性回退全局库，api 层访问形状不变）
var Service = new(ServiceGroup)

// ServiceGroup 服务组（保持既有 Service.Info 访问形状）
type ServiceGroup struct{ Info InfoService }
