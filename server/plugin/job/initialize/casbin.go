// Package initialize 白泽批量作业授权种子
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/utils"
	"go.uber.org/zap"
)

// Casbin 注册批量作业策略（幂等）：888 全量；9528 仅查询（list/detail）
func Casbin(ctx context.Context) {
	e := utils.GetCasbin()
	if e == nil {
		zap.L().Warn("job 插件：casbin 未初始化，跳过策略注册")
		return
	}
	policies := []struct {
		Path   string
		Method string
	}{
		{"/job/exec", "POST"},
		{"/job/exec/cancel", "POST"},
		{"/job/exec/list", "POST"},
		{"/job/exec/detail", "GET"},
	}
	for _, p := range policies {
		roles := []string{"888"}
		if p.Path == "/job/exec/list" || p.Path == "/job/exec/detail" {
			roles = append(roles, "9528") // 只读角色可看批次与结果，不可发起/取消
		}
		for _, role := range roles {
			has, err := e.HasPolicy(role, p.Path, p.Method)
			if err != nil || has {
				continue
			}
			if _, err := e.AddPolicy(role, p.Path, p.Method); err != nil {
				zap.L().Error("job 插件：添加 casbin 策略失败", zap.Error(err))
			}
		}
	}
}
