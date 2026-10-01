// Package model 白泽资产中心数据模型（M1）
// ER 关系：AssetRoom 1-N AssetRack；AssetHost N-N AssetProductLine；AssetHost N-1 AssetRoom/AssetRack
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// AssetStatus 主机资产状态
type AssetStatus string

const (
	AssetStatusRunning  AssetStatus = "运行中"
	AssetStatusStopped  AssetStatus = "停机"
	AssetStatusMaintain AssetStatus = "维护中"
	AssetStatusRetired  AssetStatus = "已报废"
)

// AssetHost 主机资产（物理机/云主机/虚拟机统一台账）
type AssetHost struct {
	global.GVA_MODEL
	Hostname     string             `json:"hostname" gorm:"comment:主机名" binding:"required"`
	IP           string             `json:"ip" gorm:"comment:内网IP;unique" binding:"required"`
	OS           string             `json:"os" gorm:"comment:操作系统"`
	OSVersion    string             `json:"osVersion" gorm:"comment:系统版本"`
	CPUCores     int                `json:"cpuCores" gorm:"comment:CPU核数"`
	MemGB        int                `json:"memGb" gorm:"comment:内存GB"`
	DiskGB       int                `json:"diskGb" gorm:"comment:磁盘GB"`
	SN           string             `json:"sn" gorm:"comment:序列号"`
	Vendor       string             `json:"vendor" gorm:"comment:厂商/云平台"`
	PublicIP     string             `json:"publicIp" gorm:"comment:公网IP"`
	RoomID       *uint              `json:"roomId" gorm:"comment:机房ID"`
	RackID       *uint              `json:"rackId" gorm:"comment:机柜ID"`
	UPos         string             `json:"uPos" gorm:"comment:机柜U位"`
	Status       AssetStatus        `json:"status" gorm:"comment:状态;default:运行中"`
	Owner        string             `json:"owner" gorm:"comment:负责人"`
	ExpireDate   *time.Time         `json:"expireDate" gorm:"comment:到期时间"`
	Notes        string             `json:"notes" gorm:"type:text;comment:备注"`
	ProductLines []AssetProductLine `json:"productLines" gorm:"many2many:asset_host_product_lines;"`
	Room         *AssetRoom         `json:"room" gorm:"foreignKey:RoomID"`
	Rack         *AssetRack         `json:"rack" gorm:"foreignKey:RackID"`
}

// AssetRoom 机房
type AssetRoom struct {
	global.GVA_MODEL
	Name    string `json:"name" gorm:"comment:机房名称;unique" binding:"required"`
	Region  string `json:"region" gorm:"comment:区域"`
	Address string `json:"address" gorm:"comment:详细地址"`
	Notes   string `json:"notes" gorm:"type:text;comment:备注"`
}

// AssetRack 机柜
type AssetRack struct {
	global.GVA_MODEL
	RoomID uint       `json:"roomId" gorm:"comment:机房ID" binding:"required"`
	Name   string     `json:"name" gorm:"comment:机柜名称" binding:"required"`
	TotalU int        `json:"totalU" gorm:"comment:总U数"`
	Notes  string     `json:"notes" gorm:"type:text;comment:备注"`
	Room   *AssetRoom `json:"room" gorm:"foreignKey:RoomID"`
}

// AssetProductLine 产品线
type AssetProductLine struct {
	global.GVA_MODEL
	Name  string      `json:"name" gorm:"comment:产品线名称;unique" binding:"required"`
	Owner string      `json:"owner" gorm:"comment:负责人"`
	Level string      `json:"level" gorm:"comment:等级"`
	Notes string      `json:"notes" gorm:"type:text;comment:备注"`
	Hosts []AssetHost `json:"hosts" gorm:"many2many:asset_host_product_lines;"`
}
