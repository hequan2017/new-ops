// Package initialize 白泽流水线授权种子
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/utils"
	"go.uber.org/zap"
)

// Casbin 注册流水线策略（幂等）：888 全量；9528 只读（list/find）
func Casbin(ctx context.Context) {
	e := utils.GetCasbin()
	if e == nil {
		zap.L().Warn("pipeline 插件：casbin 未初始化，跳过策略注册")
		return
	}
	policies := []struct {
		Path   string
		Method string
	}{
		{"/pipeline", "POST"},
		{"/pipeline", "PUT"},
		{"/pipeline", "DELETE"},
		{"/pipeline/list", "GET"},
		{"/pipeline/find", "GET"},
		{"/pipeline/build/start", "POST"},
		{"/pipeline/build/restart", "POST"},
		{"/pipeline/build/cancel", "POST"},
		{"/pipeline/build/approve", "POST"},
		{"/pipeline/build/list", "POST"},
		{"/pipeline/build/logs", "GET"},
		{"/sse/pipeline/build/logs", "GET"},
	}
	for _, p := range policies {
		roles := []string{"888"}
		if p.Method == "GET" {
			roles = append(roles, "9528") // 只读角色可看定义，不可改
		}
		for _, role := range roles {
			has, err := e.HasPolicy(role, p.Path, p.Method)
			if err != nil || has {
				continue
			}
			if _, err := e.AddPolicy(role, p.Path, p.Method); err != nil {
				zap.L().Error("pipeline 插件：添加 casbin 策略失败", zap.Error(err))
			}
		}
	}
}
