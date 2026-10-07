// Package model k8s 插件数据模型：命名空间级授权（三级 RBAC 的 ns 层）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// K8sNsGrant 命名空间授权：指定用户在指定集群仅可见/可读授权的 namespace
// （888 全局全量不经此表；写操作与终端仍按 casbin 仅 888——本表只做浏览面隔离）
type K8sNsGrant struct {
	global.GVA_MODEL
	ClusterID uint   `json:"clusterId" gorm:"index:idx_ns_grant,unique;comment:集群ID"`
	Namespace string `json:"namespace" gorm:"index:idx_ns_grant,unique;size:253;comment:命名空间"`
	UserID    uint   `json:"userId" gorm:"index:idx_ns_grant,unique;comment:用户ID"`
}

// TableName 表名
func (K8sNsGrant) TableName() string { return "k8s_ns_grant" }
