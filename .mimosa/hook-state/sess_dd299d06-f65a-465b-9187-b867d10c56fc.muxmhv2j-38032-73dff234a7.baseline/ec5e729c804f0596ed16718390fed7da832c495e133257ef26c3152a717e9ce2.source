// Package model 白泽 GPU 算力数据模型（M5）
// 防超卖核心：节点资源总量 - Σ(运行中实例分配量) = 可用量；开通在事务内行锁校验。
// 容器实际创建联动 container 插件（后续场次），当前交付资源台账与配额引擎。
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// 节点状态
const (
	NodeOnline  = "在线"
	NodeOffline = "离线"
)

// 实例状态
const (
	InstRunning  = "运行中"
	InstFailed   = "创建失败"
	InstReleased = "已销毁"
)

// GpuNode 算力节点
type GpuNode struct {
	global.GVA_MODEL
	Name      string `json:"name" gorm:"comment:节点名称;unique" binding:"required"`
	AssetID   *uint  `json:"assetId" gorm:"comment:关联资产ID"`
	GpuModel  string `json:"gpuModel" gorm:"comment:显卡型号" binding:"required"`
	GpuTotal  int    `json:"gpuTotal" gorm:"comment:GPU卡数" binding:"required"`
	CpuTotal  int    `json:"cpuTotal" gorm:"comment:CPU核数"`
	MemTotalGb int   `json:"memTotalGb" gorm:"comment:内存GB"`
	Status    string `json:"status" gorm:"comment:状态;default:在线"`
	Notes     string `json:"notes" gorm:"type:text;comment:备注"`
}

// GpuSpec 产品规格（定价）
type GpuSpec struct {
	global.GVA_MODEL
	Name         string  `json:"name" gorm:"comment:规格名称;unique" binding:"required"`
	GpuCount     int     `json:"gpuCount" gorm:"comment:GPU卡数" binding:"required"`
	CpuCores     int     `json:"cpuCores" gorm:"comment:CPU核数"`
	MemGb        int     `json:"memGb" gorm:"comment:内存GB"`
	PricePerHour float64 `json:"pricePerHour" gorm:"comment:单价(元/小时)"`
	Description  string  `json:"description" gorm:"comment:描述"`
}

// GpuInstance GPU 实例（资源分配台账）
type GpuInstance struct {
	global.GVA_MODEL
	NodeID       uint       `json:"nodeId" gorm:"comment:节点ID;index"`
	NodeName     string     `json:"nodeName" gorm:"comment:节点名快照"`
	SpecID       uint       `json:"specId" gorm:"comment:规格ID"`
	SpecName     string     `json:"specName" gorm:"comment:规格名快照"`
	Name         string     `json:"name" gorm:"comment:实例名称" binding:"required"`
	GpuAllocated int        `json:"gpuAllocated" gorm:"comment:分配GPU数"`
	CpuAllocated int        `json:"cpuAllocated" gorm:"comment:分配CPU核"`
	MemAllocGb   int        `json:"memAllocGb" gorm:"comment:分配内存GB"`
	Status       string     `json:"status" gorm:"comment:状态;index"`
	Creator      string     `json:"creator" gorm:"comment:开通人"`
	UserID       uint       `json:"userId" gorm:"comment:开通人ID"`
	ReleasedAt   *time.Time `json:"releasedAt" gorm:"comment:销毁时间"`
}
