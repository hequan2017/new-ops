// Package model eduadmin 教育管理一模通管：数据同步接口（需求 3.6-29~33）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduSyncAdapter 同步适配器配置（凭据 AES-GCM 密文落库，不回显明文）
type EduSyncAdapter struct {
	global.GVA_MODEL
	Source         string `json:"source" gorm:"type:varchar(100);index;comment:来源系统(人事/财务/资产/后勤/教学科研)"`
	Protocol       string `json:"protocol" gorm:"type:varchar(32);comment:HTTP/WebService/FTP"`
	Endpoint       string `json:"endpoint" gorm:"type:varchar(500);comment:拉取端点"`
	CredentialEnc  string `json:"-" gorm:"type:varchar(1000);comment:凭据密文(AES-GCM)"`
	SyncFreqMinutes int   `json:"syncFreqMinutes" gorm:"default:5;comment:同步频率(分钟,≤5 口径)"`
	Status         string `json:"status" gorm:"type:varchar(16);comment:启用/停用"`
}

func (EduSyncAdapter) TableName() string { return "edu_sync_adapter" }

// EduSyncTask 同步任务执行记录
type EduSyncTask struct {
	global.GVA_MODEL
	AdapterId uint   `json:"adapterId" gorm:"index;comment:适配器ID"`
	Pulled    int    `json:"pulled" gorm:"comment:拉取行数"`
	Inserted  int    `json:"inserted" gorm:"comment:新入库行数"`
	Skipped   int    `json:"skipped" gorm:"comment:唯一键去重跳过数"`
	Failed    int    `json:"failed" gorm:"comment:校验失败行数"`
	OK        bool   `json:"ok" gorm:"comment:是否成功"`
	Message   string `json:"message" gorm:"type:varchar(500);comment:执行说明"`
}

func (EduSyncTask) TableName() string { return "edu_sync_task" }
