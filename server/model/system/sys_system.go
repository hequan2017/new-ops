package system

import (
	"github.com/hequan2017/new-ops/server/config"
)

// 配置文件结构体
type System struct {
	Config config.Server `json:"config"`
}
