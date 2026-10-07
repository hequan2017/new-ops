// Package model 白泽批量作业数据模型（M2：批量命令执行）
// 批次（JobExecRecord）与每主机结果（JobExecResult）一对多；
// 脚本库/变量组（job_script/job_variable_group）随后续场次补齐。
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// 批次状态
const (
	ExecStatusRunning   = "执行中"
	ExecStatusFinished  = "已完成"
	ExecStatusCanceled  = "已取消"
)

// 主机执行结果状态
const (
	ResultSuccess = "成功"
	ResultFailed  = "失败"
	ResultTimeout = "超时"
	ResultCanceled = "取消"
	ResultSkipped = "跳过"
)

// JobExecRecord 批量执行批次
type JobExecRecord struct {
	global.GVA_MODEL
	Command     string     `json:"command" gorm:"type:text;comment:执行的命令"`
	Concurrency int        `json:"concurrency" gorm:"comment:并发数"`
	TimeoutSec  int        `json:"timeoutSec" gorm:"comment:单机超时秒"`
	CredentialID *uint     `json:"credentialId" gorm:"comment:统一凭据ID(空则用主机绑定凭据)"`
	VariableGroupID *uint  `json:"variableGroupId" gorm:"comment:显式变量组ID(空则按主机资产组自动注入)"`
	Total       int        `json:"total" gorm:"comment:主机总数"`
	Success     int        `json:"success" gorm:"comment:成功数"`
	Failed      int        `json:"failed" gorm:"comment:失败数"`
	Status      string     `json:"status" gorm:"comment:状态;default:执行中;index"`
	Operator    string     `json:"operator" gorm:"comment:操作人"`
	UserID      uint       `json:"userId" gorm:"comment:操作用户ID;index"`
	StartedAt   time.Time  `json:"startedAt" gorm:"comment:开始时间"`
	FinishedAt  *time.Time `json:"finishedAt" gorm:"comment:结束时间"`
}

// JobExecResult 单主机执行结果
type JobExecResult struct {
	global.GVA_MODEL
	RecordID  uint   `json:"recordId" gorm:"comment:批次ID;index"`
	HostID    uint   `json:"hostId" gorm:"comment:主机ID"`
	Hostname  string `json:"hostname" gorm:"comment:主机名快照"`
	Command   string `json:"command" gorm:"type:text;comment:该主机实际执行的命令(变量渲染后)"`
	IP        string `json:"ip" gorm:"comment:IP快照"`
	Status    string `json:"status" gorm:"comment:结果状态"`
	Output    string `json:"output" gorm:"type:text;comment:输出"`
	Error     string `json:"error" gorm:"type:text;comment:错误信息"`
	ElapsedMs int64  `json:"elapsedMs" gorm:"comment:耗时毫秒"`
}
