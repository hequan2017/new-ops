// Package model artsci 学科工具集：执行留痕
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// ArtsciRunLog 工具执行留痕（成败均落库）
type ArtsciRunLog struct {
	global.GVA_MODEL
	ToolId     uint   `json:"toolId" gorm:"index;comment:工具ID"`
	Name       string `json:"name" gorm:"type:varchar(128);comment:工具名"`
	ArgsJSON   string `json:"argsJSON" gorm:"type:text;comment:入参"`
	ResultJSON string `json:"resultJSON" gorm:"type:mediumtext;comment:出参"`
	Ok         bool   `json:"ok" gorm:"index;comment:是否成功"`
	LatencyMs  int64  `json:"latencyMs" gorm:"comment:耗时毫秒"`
	ErrMsg     string `json:"errMsg" gorm:"type:varchar(500);comment:错误"`
}

func (ArtsciRunLog) TableName() string { return "artsci_run_log" }
