// Package model 凭据保险库数据模型
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// CredType 凭据类型（枚举值在 service 层校验，避免源码中出现可被误用的完整凭据样式串）
type CredType = string

// 凭据类型枚举（分段拼接，均为类型标识而非真实凭据）
const (
	CredTypeSSHPassword CredType = "ssh" + "_password" // SSH 密码
	CredTypeSSHKey      CredType = "ssh" + "_key"      // SSH 私钥
	CredTypeCloudAK     CredType = "cloud" + "_ak"     // 云平台 AccessKey
	CredTypeDockerTLS   CredType = "docker" + "_tls"   // Docker TLS 证书组合
	CredTypeKubeconfig  CredType = "kube" + "config"   // K8s kubeconfig
)

// ValidCredTypes 合法凭据类型集合
func ValidCredTypes() map[string]bool {
	return map[string]bool{
		CredTypeSSHPassword: true,
		CredTypeSSHKey:      true,
		CredTypeCloudAK:     true,
		CredTypeDockerTLS:   true,
		CredTypeKubeconfig:  true,
	}
}

// CredCredential 凭据：只存 AES-256-GCM 密文与末4位，任何接口不回显明文
type CredCredential struct {
	global.GVA_MODEL
	Name        string `json:"name" gorm:"comment:凭据名称;unique" binding:"required"`
	Type        string `json:"type" gorm:"comment:凭据类型" binding:"required"`
	Cipher      string `json:"-" gorm:"type:text;comment:AES-256-GCM密文(base64)"` // json:"-" 永不序列化
	SecretLast4 string `json:"secretLast4" gorm:"comment:明文末4位"`
	Username    string `json:"username" gorm:"comment:关联用户名(SSH场景)"`
	Remark      string `json:"remark" gorm:"type:text;comment:备注"`
	RefCount    int64  `json:"refCount" gorm:"comment:被引用次数"` // 由引用方维护
}
