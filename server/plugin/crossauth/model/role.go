// Package model crossauth 跨校访问认证与安全：角色模板（需求 3.7-5）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CaRoleTemplate 跨校角色模板（内置基础权限 JSON）
type CaRoleTemplate struct {
	global.GVA_MODEL
	RoleName    string `json:"roleName" gorm:"type:varchar(100);uniqueIndex;comment:角色名"`
	Category    string `json:"category" gorm:"type:varchar(32);index;comment:教学/科研/实训/管理/企业/审计/运维"`
	PermJSON    string `json:"permJSON" gorm:"type:text;comment:内置基础权限(JSON)"`
	Description string `json:"description" gorm:"type:varchar(500);comment:角色说明"`
}

func (CaRoleTemplate) TableName() string { return "ca_role_template" }
