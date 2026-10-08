// Package model 白泽容器管理：Compose 编排项目（M8 C2）
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// Compose 项目状态
const (
	ComposeStatusRunning = "运行中"
	ComposeStatusPartial = "部分运行"
	ComposeStatusDown    = "已停止"
	ComposeStatusUnknown = "未知"
)

// DockerComposeProject Compose 编排项目（compose 文件内容留存，up/down/ps 均按此内容执行）
type DockerComposeProject struct {
	global.GVA_MODEL
	Name       string     `json:"name" gorm:"comment:项目名(compose -p);unique" binding:"required"`
	EndpointID uint       `json:"endpointId" gorm:"comment:Docker接入点ID;index"`
	Content    string     `json:"content" gorm:"type:longtext;comment:compose 文件内容"`
	Status     string     `json:"status" gorm:"comment:状态;default:未知;index"`
	Services   string     `json:"services" gorm:"type:text;comment:最近一次 ps 结果(JSON)"`
	LastOpAt   *time.Time `json:"lastOpAt" gorm:"comment:最近操作时间"`
	CreatedBy  string     `json:"createdBy" gorm:"comment:创建人"`
	Notes      string     `json:"notes" gorm:"type:text;comment:备注"`
}
