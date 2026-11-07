// Package model oplib：执行留痕
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// OplibRunLog 算子执行留痕（H04 编排复用为节点日志）
type OplibRunLog struct {
	global.GVA_MODEL
	OperatorId uint   `json:"operatorId" gorm:"index;comment:算子ID"`
	Name       string `json:"name" gorm:"type:varchar(128);comment:算子名"`
	ArgsJSON   string `json:"argsJSON" gorm:"type:text;comment:入参"`
	ResultJSON string `json:"resultJSON" gorm:"type:mediumtext;comment:出参"`
	Ok         bool   `json:"ok" gorm:"index;comment:是否成功"`
	LatencyMs  int64  `json:"latencyMs" gorm:"comment:耗时毫秒"`
	ErrMsg     string `json:"errMsg" gorm:"type:varchar(500);comment:错误"`
}

func (OplibRunLog) TableName() string { return "oplib_run_log" }
