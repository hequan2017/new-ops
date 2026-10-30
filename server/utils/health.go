package utils

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/global"
)

// processStart 进程启动时刻：本包随路由初始化在启动期加载，以包级变量初始化时间近似进程启动时间
var processStart = time.Now()

// HealthDetail 免鉴权健康检查详情，供监控探针与部署冒烟脚本使用。
// 返回字段：status 固定 ok；app 应用名；version 平台发布版本；base_version 底座版本；
// uptime 进程运行秒数；go_version Go 运行时版本；timestamp 服务器当前时间。
func HealthDetail(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"app":          global.AppName,
		"version":      global.ReleaseVersion,
		"base_version": global.Version,
		"uptime":       int64(time.Since(processStart).Seconds()),
		"go_version":   runtime.Version(),
		"timestamp":    time.Now().Format(time.DateTime),
	})
}
