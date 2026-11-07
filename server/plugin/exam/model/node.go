// Package model exam：考试节点与领卷分片
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// ExamNode 考试节点（采购件③ 分布式：终端/考点接入，心跳保活）
type ExamNode struct {
	global.GVA_MODEL
	NodeKey       string    `json:"nodeKey" gorm:"type:varchar(64);uniqueIndex;comment:节点标识(机器码)"`
	Name          string    `json:"name" gorm:"type:varchar(64);comment:节点名"`
	Region        string    `json:"region" gorm:"type:varchar(64);index;comment:区域/考点"`
	Load          int       `json:"load" gorm:"default:0;comment:当前负载(在考人数)"`
	LastHeartbeat time.Time `json:"lastHeartbeat" gorm:"comment:最后心跳"`
	Status        string    `json:"status" gorm:"type:varchar(16);default:online;index;comment:online/offline"`
}

func (ExamNode) TableName() string { return "exam_node" }

// ExamShard 领卷分片（考生→节点分配；sheetId 唯一=一卷一片幂等）
type ExamShard struct {
	global.GVA_MODEL
	ExamId   uint      `json:"examId" gorm:"index;comment:考试ID"`
	SheetId  uint      `json:"sheetId" gorm:"uniqueIndex;comment:答卷ID"`
	NodeKey  string    `json:"nodeKey" gorm:"type:varchar(64);index;comment:分配节点"`
	IssuedAt time.Time `json:"issuedAt" gorm:"comment:发放时间"`
}

func (ExamShard) TableName() string { return "exam_shard" }
