// Package model 白泽 Agent 通道：实例注册表（M9 A1）
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// AgentInstance Agent 实例（一机一 Agent；在线状态按心跳时效计算，不落库）
type AgentInstance struct {
	global.GVA_MODEL
	AssetID       uint       `json:"assetId" gorm:"comment:资产ID;uniqueIndex"`
	TokenHash     string     `json:"-" gorm:"size:96;comment:接入令牌哈希(SHA256)"` // 永不序列化
	TokenHint     string     `json:"tokenHint" gorm:"size:8;comment:令牌末4位"`
	Version       string     `json:"version" gorm:"size:32;comment:Agent版本"`
	OsArch        string     `json:"osArch" gorm:"size:32;comment:os/arch"`
	Hostname      string     `json:"hostname" gorm:"size:128;comment:上报主机名"`
	Labels        string     `json:"labels" gorm:"size:255;comment:标签(逗号分隔)"`
	LastHeartbeat *time.Time `json:"lastHeartbeat" gorm:"comment:最近心跳"`
}
