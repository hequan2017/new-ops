// Package initialize 白泽容器管理授权种子
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/utils"
	"go.uber.org/zap"
)

// Casbin 注册容器管理策略（幂等）：888 全量；9528 只读（list）
func Casbin(ctx context.Context) {
	e := utils.GetCasbin()
	if e == nil {
		zap.L().Warn("container 插件：casbin 未初始化，跳过策略注册")
		return
	}
	policies := []struct {
		Path   string
		Method string
	}{
		{"/container/endpoint", "POST"},
		{"/container/endpoint", "PUT"},
		{"/container/endpoint", "DELETE"},
		{"/container/endpoint/list", "GET"},
		{"/container/endpoint/check", "POST"},
		{"/container/container", "POST"},
		{"/container/container/list", "GET"},
		{"/container/container/action", "POST"},
	}
	for _, p := range policies {
		roles := []string{"888"}
		if p.Path == "/container/endpoint/list" {
			roles = append(roles, "9528")
		}
		for _, role := range roles {
			has, err := e.HasPolicy(role, p.Path, p.Method)
			if err != nil || has {
				continue
			}
			if _, err := e.AddPolicy(role, p.Path, p.Method); err != nil {
				zap.L().Error("container 插件：添加 casbin 策略失败", zap.Error(err))
			}
		}
	}
}
