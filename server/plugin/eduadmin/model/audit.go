// Package model eduadmin 教育管理一模通管：审计知识语料（需求 3.6-37）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// EduAuditItem 审计条目（领域/风险要点/内控建议）
type EduAuditItem struct {
	global.GVA_MODEL
	Domain       string `json:"domain" gorm:"type:varchar(32);index;comment:财务/采购/人事/资产/工程/数据安全"`
	Title        string `json:"title" gorm:"type:varchar(300);comment:条目标题"`
	RiskPoints   string `json:"riskPoints" gorm:"type:text;comment:风险要点"`
	ControlAdvice string `json:"controlAdvice" gorm:"type:text;comment:内控建议"`
}

func (EduAuditItem) TableName() string { return "edu_audit_item" }
