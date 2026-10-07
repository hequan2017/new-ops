// Package model 白泽容器事件日志（M4 C1：docker events 区间拉取落库）
package model

import (
	"time"

	"github.com/hequan2017/new-ops/server/global"
)

// DockerEventLog 容器事件（30s 巡检循环按 Since/Until 区间拉取）
type DockerEventLog struct {
	global.GVA_MODEL
	EndpointID  uint      `json:"endpointId" gorm:"comment:接入点ID;index"`
	Action      string    `json:"action" gorm:"comment:动作(start/die/destroy等)"`
	ContainerID string    `json:"containerId" gorm:"size:64;comment:容器ID(12位)"`
	ContainerNm string    `json:"containerName" gorm:"comment:容器名"`
	OccurredAt  time.Time `json:"occurredAt" gorm:"comment:事件发生时间"`
}
