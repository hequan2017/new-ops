// Package model 白泽 Kubernetes 多集群管理数据模型（M4 K1）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// 集群状态
const (
	K8sStatusOnline  = "在线"
	K8sStatusOffline = "离线"
	K8sStatusUnknown = "未知"
)

// K8sCluster 集群注册（kubeconfig AES-256-GCM 加密落库）
type K8sCluster struct {
	global.GVA_MODEL
	Name             string `json:"name" gorm:"comment:集群名称;unique" binding:"required"`
	Remark           string `json:"remark" gorm:"type:text;comment:备注"`
	KubeconfigCipher string `json:"-" gorm:"type:text;comment:kubeconfig密文(base64)"` // json:"-" 永不序列化
	Server           string `json:"server" gorm:"comment:API Server地址"`
	Version          string `json:"version" gorm:"comment:集群版本"`
	NodeCount        int    `json:"nodeCount" gorm:"comment:节点数"`
	Status           string `json:"status" gorm:"comment:状态;default:未知"`
}
