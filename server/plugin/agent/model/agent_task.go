// Package model 白泽 Agent 通道：任务留痕（M9 A2）
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// AgentTask 下发任务留痕（exec 类型；任务号幂等去重键，结果/超时全量留档）
type AgentTask struct {
	global.GVA_MODEL
	TaskNo     string     `json:"taskNo" gorm:"size:64;uniqueIndex;comment:任务号(幂等去重键)"`
	AssetID    uint       `json:"assetId" gorm:"index;comment:资产ID"`
	Type       string     `json:"type" gorm:"size:32;comment:任务类型(exec)"`
	Command    string     `json:"command" gorm:"type:text;comment:命令内容"`
	Status     string     `json:"status" gorm:"size:16;index;comment:状态(成功/失败/超时)"`
	ExitCode   int        `json:"exitCode" gorm:"comment:退出码(-1=执行异常)"`
	Output     string     `json:"output" gorm:"type:text;comment:输出(截断留存)"`
	ErrMsg     string     `json:"errMsg" gorm:"size:512;comment:错误信息"`
	ElapsedMs  int64      `json:"elapsedMs" gorm:"comment:耗时毫秒"`
	Operator   string     `json:"operator" gorm:"size:64;comment:触发人"`
	FinishedAt *time.Time `json:"finishedAt" gorm:"comment:完成时间"`
}
