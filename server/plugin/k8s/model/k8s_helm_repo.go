// Package model k8s 插件数据模型：Helm chart 仓库（全局配置，非按集群）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// K8sHelmRepo Helm chart 仓库登记（http/https 仓库地址，安装时按 repo 引用免上传 tgz）
type K8sHelmRepo struct {
	global.GVA_MODEL
	Name   string `json:"name" gorm:"index:idx_helm_repo_name,unique;comment:仓库名"`
	URL    string `json:"url" gorm:"comment:仓库地址"`
	Remark string `json:"remark" gorm:"comment:备注"`
}

// TableName 表名
func (K8sHelmRepo) TableName() string { return "k8s_helm_repo" }
