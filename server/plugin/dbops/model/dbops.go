// Package model 白泽数据库工单数据模型（M6 dbops）
// 实例密码 AES-256-GCM 加密落库（复用 asset/crypto），接口不回显明文；
// goInception 审核引擎未接入前，SQL 工单仅支持创建/取消/明细（审核动作明确报未配置）。
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// 实例状态
const (
	InstanceOnline  = "在线"
	InstanceOffline = "离线"
	InstanceUnknown = "未知"
)

// 工单状态
const (
	OrderPending   = "待审核"
	OrderApproved  = "审核通过"
	OrderExecuting = "执行中"
	OrderSuccess   = "成功"
	OrderFailed    = "失败"
	OrderCanceled  = "已取消"
)

// DbopsInstance MySQL 实例
type DbopsInstance struct {
	global.GVA_MODEL
	Name         string     `json:"name" gorm:"comment:实例名称;unique" binding:"required"`
	Host         string     `json:"host" gorm:"comment:主机" binding:"required"`
	Port         int        `json:"port" gorm:"comment:端口;default:3306"`
	Username     string     `json:"username" gorm:"comment:账号"`
	PasswordEnc  string     `json:"-" gorm:"size:255;comment:密码密文(AES-GCM)"`
	Database     string     `json:"database" gorm:"comment:默认库"`
	Status       string     `json:"status" gorm:"comment:状态;default:未知;index"`
	Version      string     `json:"version" gorm:"comment:版本"`
	LastCheckAt  *time.Time `json:"lastCheckAt" gorm:"comment:最近检测时间"`
	Notes        string     `json:"notes" gorm:"type:text;comment:备注"`
}

// DbopsOrder SQL 上线工单
type DbopsOrder struct {
	global.GVA_MODEL
	InstanceID   uint       `json:"instanceId" gorm:"comment:实例ID;index"`
	InstanceNm   string     `json:"instanceName" gorm:"comment:实例名快照"`
	Title        string     `json:"title" gorm:"comment:工单标题"`
	AuditPayload string     `json:"sqlText" gorm:"column:sql_text;type:longtext;comment:SQL文本(送审载荷)"`
	Status       string     `json:"status" gorm:"comment:状态;default:待审核;index"`
	AuditResult  string     `json:"auditResult" gorm:"type:text;comment:审核结果JSON"`
	ExecResult   string     `json:"execResult" gorm:"type:text;comment:执行结果JSON"`
	Creator      string     `json:"creator" gorm:"comment:发起人"`
	UserID       uint       `json:"userId" gorm:"comment:发起人ID;index"`
	FinishedAt   *time.Time `json:"finishedAt" gorm:"comment:结束时间"`
}
