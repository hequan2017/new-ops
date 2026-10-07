// Package model 资产组（数据权限单元）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// AssetGroup 资产组：绑定主机与用户，非超管用户仅可见其所在组的主机
type AssetGroup struct {
	global.GVA_MODEL
	Name  string `json:"name" gorm:"comment:资产组名称;unique" binding:"required"`
	Notes string `json:"notes" gorm:"type:text;comment:备注"`
	// HostIDs/UserIDs 由 service 层维护关联表（列表返回），不走 gorm 关联
}

// AssetGroupHost 资产组-主机关联
type AssetGroupHost struct {
	GroupID uint `json:"groupId" gorm:"comment:资产组ID;primaryKey"`
	HostID  uint `json:"hostId" gorm:"comment:主机ID;primaryKey"`
}

// TableName 关联表名
func (AssetGroupHost) TableName() string { return "asset_group_hosts" }

// AssetGroupUser 资产组-用户关联
type AssetGroupUser struct {
	GroupID uint   `json:"groupId" gorm:"comment:资产组ID;primaryKey"`
	UserID  uint   `json:"userId" gorm:"comment:用户ID;primaryKey"`
	Username string `json:"username" gorm:"comment:用户名（冗余，便于展示）"`
}

// TableName 关联表名
func (AssetGroupUser) TableName() string { return "asset_group_users" }
