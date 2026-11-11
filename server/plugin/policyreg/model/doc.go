// Package model policyreg 高校制度大模型：制度导入（需求 3.8-1/2）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// PolicyDoc 制度文档（密级权限范围 + SM4 存取口径标注 + 涉密条款脱敏稿）
type PolicyDoc struct {
	global.GVA_MODEL
	Name          string `json:"name" gorm:"type:varchar(300);comment:制度名称"`
	Format        string `json:"format" gorm:"type:varchar(16);index;comment:Word/PDF/HTML/TXT/MD"`
	RawText       string `json:"rawText" gorm:"type:mediumtext;comment:原文文本(二进制格式编目留空,口径同 A03)"`
	SecurityLevel string `json:"securityLevel" gorm:"type:varchar(16);index;comment:公开/内部/涉密"`
	PermScope     string `json:"permScope" gorm:"type:varchar(500);comment:权限范围(角色/单位,逗号分隔)"`
	Sm4Mark       bool   `json:"sm4Mark" gorm:"comment:涉密字段 SM4 加密存取口径"`
	MaskedText    string `json:"maskedText" gorm:"type:mediumtext;comment:自动脱敏稿(涉密条款替换)"`
	ContentHash   string `json:"contentHash" gorm:"type:char(64);uniqueIndex;comment:SHA256(name|rawText) 幂等键"`
}

func (PolicyDoc) TableName() string { return "policy_doc" }
