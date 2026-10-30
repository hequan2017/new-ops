package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/global"
	"go.uber.org/zap"
)

// defaultSlowRequestSec 配置未设置（0）时采用的默认慢请求阈值
const defaultSlowRequestSec = 2

// SlowRequest 慢请求标记：耗时超过配置阈值（system.slow-request-threshold，秒）的请求
// 以 warn 级别落 zap 日志。release 模式下不挂 gin.Logger，常规请求无日志，靠此暴露慢接口；
// 阈值配置负数表示关闭。
func SlowRequest() gin.HandlerFunc {
	return slowRequest(resolveSlowThreshold(global.GVA_CONFIG.System.SlowRequestThreshold))
}

// resolveSlowThreshold 解析配置的整秒阈值：0 用默认 2s，负数返回 0 表示关闭
func resolveSlowThreshold(sec int) time.Duration {
	switch {
	case sec < 0:
		return 0
	case sec == 0:
		return defaultSlowRequestSec * time.Second
	default:
		return time.Duration(sec) * time.Second
	}
}

// slowRequest 阈值 <=0 时返回直通中间件（关闭态）
func slowRequest(threshold time.Duration) gin.HandlerFunc {
	if threshold <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		if cost := time.Since(start); cost > threshold {
			global.GVA_LOG.Warn("slow request",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.String("ip", c.ClientIP()),
				zap.Duration("cost", cost),
			)
		}
	}
}
