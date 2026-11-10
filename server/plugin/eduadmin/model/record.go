// Package model eduadmin 教育管理一模通管：全场景数据接入（需求 3.6-1）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduRecord 全场景数据记录（四域：人事/财务/资产/后勤；外部唯一键幂等去重）
type EduRecord struct {
	global.GVA_MODEL
	Domain      string `json:"domain" gorm:"type:varchar(16);uniqueIndex:idx_edu_domain_ext;comment:人事/财务/资产/后勤"`
	ExternalKey string `json:"externalKey" gorm:"type:varchar(200);uniqueIndex:idx_edu_domain_ext;comment:外部系统唯一键"`
	PayloadJSON string `json:"payloadJSON" gorm:"type:text;comment:业务负载(JSON 对象)"`
	Source      string `json:"source" gorm:"type:varchar(100);comment:来源系统(人事系统/财务系统/资产系统/后勤系统)"`
}

func (EduRecord) TableName() string { return "edu_record" }
