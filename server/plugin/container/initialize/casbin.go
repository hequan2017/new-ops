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
		{"/container/event/list", "GET"},
		{"/container/image/list", "GET"},
		{"/container/image/pull", "POST"},
		{"/container/image/pull-status", "GET"},
		{"/container/image", "DELETE"},
		{"/container/image/tag", "POST"},
		{"/container/image/export", "GET"},
		{"/container/image/import", "POST"},
		{"/container/network/list", "GET"},
		{"/container/network", "POST"},
		{"/container/network", "DELETE"},
		{"/container/volume/list", "GET"},
		{"/container/volume", "DELETE"},
		{"/container/container/stats", "GET"},
		{"/container/container/logws", "GET"},
		{"/container/container/execws", "GET"},
		{"/container/compose", "POST"},
		{"/container/compose", "PUT"},
		{"/container/compose", "DELETE"},
		{"/container/compose/list", "GET"},
		{"/container/compose/detail", "GET"},
		{"/container/compose/ps", "GET"},
		{"/container/compose/action", "POST"},
	}
	for _, p := range policies {
		roles := []string{"888"}
		switch p.Path {
		case "/container/endpoint/list", "/container/container/logws":
			// 只读流：9528 可看日志
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
