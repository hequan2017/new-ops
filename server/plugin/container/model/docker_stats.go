// Package model 白泽容器管理：容器资源统计采样（M8 C2 历史图表化）
package model

import (
	"github.com/hequan2017/new-ops/server/global"
)

// DockerStatsSample 容器资源统计采样点（5 分钟一轮，7 天留存，与 docker_event_log 口径一致）
type DockerStatsSample struct {
	global.GVA_MODEL
	EndpointID  uint    `json:"endpointId" gorm:"comment:接入点ID;index"`
	ContainerID string  `json:"containerId" gorm:"size:64;comment:容器ID(12位);index"`
	Name        string  `json:"name" gorm:"size:128;comment:容器名"`
	CPUPercent  float64 `json:"cpuPercent" gorm:"comment:CPU使用率"`
	MemUsedMB   float64 `json:"memUsedMb" gorm:"comment:内存使用MB"`
	MemLimitMB  float64 `json:"memLimitMb" gorm:"comment:内存上限MB"`
	MemPercent  float64 `json:"memPercent" gorm:"comment:内存使用率"`
	NetRxMB     float64 `json:"netRxMb" gorm:"comment:网络累计接收MB"`
	NetTxMB     float64 `json:"netTxMb" gorm:"comment:网络累计发送MB"`
	Pids        int     `json:"pids" gorm:"comment:进程数"`
}
