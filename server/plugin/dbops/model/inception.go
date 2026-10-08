// Package model 白泽数据库工单：goInception 审核引擎配置（M6 收官）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// DbopsInceptionConfig goInception 审核引擎（单行语义，Name=default）
// 客户端鉴权口径：连接 goInception 的用户名嵌入目标库地址（user@host:port），密码为目标库密码；
// 本表保存备份库凭据（回滚语句落 `<db>_inception_backup`）。
type DbopsInceptionConfig struct {
	global.GVA_MODEL
	Name          string `json:"name" gorm:"size:32;comment:配置名;unique"`
	Host          string `json:"host" gorm:"comment:goInception 主机"`
	Port          int    `json:"port" gorm:"comment:goInception 端口;default:4000"`
	BackupHost    string `json:"backupHost" gorm:"comment:备份库主机"`
	BackupPort    int    `json:"backupPort" gorm:"comment:备份库端口;default:3306"`
	BackupUser    string `json:"backupUser" gorm:"comment:备份库账号"`
	BackupPassEnc string `json:"-" gorm:"size:255;comment:备份库密码密文(AES-GCM)"`
	Notes         string `json:"notes" gorm:"type:text;comment:备注"`
}
