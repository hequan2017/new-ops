// Package model 白泽容器管理数据模型（M4 C1：Docker 接入点）
// 容器实时查询不落库，仅接入点（endpoint）与巡检状态持久化（DEV_PLAN 6.2）。
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// 接入点状态
const (
	EndpointOnline  = "在线"
	EndpointOffline = "离线"
	EndpointUnknown = "未知"
)

// DockerEndpoint Docker 接入点
type DockerEndpoint struct {
	global.GVA_MODEL
	Name            string     `json:"name" gorm:"comment:接入点名称;unique" binding:"required"`
	Addr            string     `json:"addr" gorm:"comment:地址(unix:///… 或 tcp://host:port)" binding:"required"`
	TLSCredentialID *uint      `json:"tlsCredentialId" gorm:"comment:TLS凭据ID(docker_tls, tcp+tls时必填)"`
	Status          string     `json:"status" gorm:"comment:巡检状态;default:未知;index"`
	DockerVersion   string     `json:"dockerVersion" gorm:"comment:Docker版本"`
	LastCheckAt     *time.Time `json:"lastCheckAt" gorm:"comment:最近巡检时间"`
	LastEventAt     *time.Time `json:"lastEventAt" gorm:"comment:事件拉取水位"`
	Notes           string     `json:"notes" gorm:"type:text;comment:备注"`
}
