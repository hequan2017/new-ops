// Package model policyreg 高校制度大模型：模型鉴权知识库（需求 3.8-7）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// PolicyPermGrant 多角色精细权限 grant（制度文本按需访问：角色×分类×最高密级）
type PolicyPermGrant struct {
	global.GVA_MODEL
	RoleName   string `json:"roleName" gorm:"type:varchar(100);uniqueIndex:idx_pr_role_cat;comment:角色"`
	DocCategory string `json:"docCategory" gorm:"type:varchar(32);uniqueIndex:idx_pr_role_cat;index;comment:教学/科研/人事/财务/资产/综合"`
	MaxLevel   string `json:"maxLevel" gorm:"type:varchar(16);comment:可访问最高密级(公开/内部/涉密)"`
	Allowed    bool   `json:"allowed" gorm:"comment:是否允许"`
}

func (PolicyPermGrant) TableName() string { return "policy_perm_grant" }
