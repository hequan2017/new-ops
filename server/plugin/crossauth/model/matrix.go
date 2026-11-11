// Package model crossauth 跨校访问认证与安全：四级权限矩阵（需求 3.7-7）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CaPermGrant 四级权限授权（校-部门-用户-场景；* 通配；最具体优先）
type CaPermGrant struct {
	global.GVA_MODEL
	School   string `json:"school" gorm:"type:varchar(100);index;comment:院校(* 通配)"`
	Dept     string `json:"dept" gorm:"type:varchar(100);index;comment:部门(* 通配)"`
	UserName string `json:"userName" gorm:"type:varchar(100);index;comment:用户(* 通配)"`
	Scene    string `json:"scene" gorm:"type:varchar(200);index;comment:场景(* 通配)"`
	Allowed  bool   `json:"allowed" gorm:"comment:是否允许"`
	Remark   string `json:"remark" gorm:"type:varchar(300);comment:备注"`
}

func (CaPermGrant) TableName() string { return "ca_perm_grant" }
