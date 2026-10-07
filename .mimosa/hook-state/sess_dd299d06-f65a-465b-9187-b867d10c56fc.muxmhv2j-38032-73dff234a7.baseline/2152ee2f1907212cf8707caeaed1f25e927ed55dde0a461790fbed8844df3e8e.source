// Package model 资产变更历史
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// 资产变更动作
const (
	HostHistoryCreate = "创建"
	HostHistoryUpdate = "更新"
	HostHistoryDelete = "删除"
)

// AssetHostHistory 主机资产变更历史
type AssetHostHistory struct {
	global.GVA_MODEL
	HostID   uint   `json:"hostId" gorm:"comment:主机ID;index"`
	Action   string `json:"action" gorm:"comment:动作(创建/更新/删除)"`
	Snapshot string `json:"snapshot" gorm:"type:text;comment:变更后快照JSON"`
	Operator string `json:"operator" gorm:"comment:操作人"`
}
