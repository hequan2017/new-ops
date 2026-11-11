// Package model crossauth 跨校访问认证与安全：权限分配双模式（需求 3.7-6）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CaAutoGrantRule 自动匹配规则（院校+部门→角色模板；dept=* 通配）
type CaAutoGrantRule struct {
	global.GVA_MODEL
	School   string `json:"school" gorm:"type:varchar(100);index;comment:所属院校"`
	Dept     string `json:"dept" gorm:"type:varchar(100);comment:部门(* 通配)"`
	RoleName string `json:"roleName" gorm:"type:varchar(100);comment:角色模板"`
}

func (CaAutoGrantRule) TableName() string { return "ca_auto_grant_rule" }

// CaGrantRequest 人工审核申请（本校+资源归属校双重审核状态机）
type CaGrantRequest struct {
	global.GVA_MODEL
	UserName       string `json:"userName" gorm:"type:varchar(100);index;comment:申请人"`
	School         string `json:"school" gorm:"type:varchar(100);comment:所属院校"`
	DesiredRole    string `json:"desiredRole" gorm:"type:varchar(100);comment:申请角色"`
	TargetResource string `json:"targetResource" gorm:"type:varchar(200);comment:目标资源"`
	Status         string `json:"status" gorm:"type:varchar(16);index;comment:提交/本校审核/归属校审核/通过/驳回"`
	HomeApprover   string `json:"homeApprover" gorm:"type:varchar(100);comment:本校审核人"`
	HostApprover   string `json:"hostApprover" gorm:"type:varchar(100);comment:归属校审核人"`
	PushLog        string `json:"pushLog" gorm:"type:varchar(500);comment:结果推送记录"`
}

func (CaGrantRequest) TableName() string { return "ca_grant_request" }

// CaGrantAssignment 权限分配结果
type CaGrantAssignment struct {
	global.GVA_MODEL
	UserName string `json:"userName" gorm:"type:varchar(100);index;comment:用户"`
	School   string `json:"school" gorm:"type:varchar(100);comment:所属院校"`
	RoleName string `json:"roleName" gorm:"type:varchar(100);index;comment:角色模板"`
	Source   string `json:"source" gorm:"type:varchar(16);comment:自动/人工"`
	PermJSON string `json:"permJSON" gorm:"type:text;comment:生效权限(JSON)"`
}

func (CaGrantAssignment) TableName() string { return "ca_grant_assignment" }
