// Package model ziqiang 自强-理工科大模型：区块链存证（需求 3.5-8）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ZqEvidence 存证块：SHA256 链式追加（BlockHash=f(height,ts,dataHash,prevHash)），
// 创世块 prevHash=64 个 0；payload 留底供完整性校验重算。
type ZqEvidence struct {
	global.GVA_MODEL
	BizKey    string `json:"bizKey" gorm:"type:varchar(200);index;comment:业务关联号(语料ID/案件号等)"`
	Payload   string `json:"payload" gorm:"type:mediumtext;comment:存证原文"`
	DataHash  string `json:"dataHash" gorm:"type:char(64);index;comment:SHA256(原文)"`
	PrevHash  string `json:"prevHash" gorm:"type:char(64);comment:上一块哈希(创世为全 0)"`
	BlockHash string `json:"blockHash" gorm:"type:char(64);uniqueIndex;comment:本块哈希"`
	Height    int64  `json:"height" gorm:"uniqueIndex;comment:块高(从 0 起)"`
	Timestamp int64  `json:"timestamp" gorm:"comment:存证时间戳(unix 秒)"`
}

func (ZqEvidence) TableName() string { return "ziqiang_evidence" }
