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
		{"/job/script", "POST"},
		{"/job/script", "PUT"},
		{"/job/script", "DELETE"},
		{"/job/script/list", "GET"},
		{"/job/script/versions", "GET"},
		{"/job/vargroup", "POST"},
		{"/job/vargroup", "PUT"},
		{"/job/vargroup", "DELETE"},
		{"/job/vargroup/list", "GET"},
	}
	for _, p := range policies {
		roles := []string{"888"}
		// 只读角色：批次查看 + 脚本/变量组列表与版本查看，不可写
		if (p.Method == "GET") ||
			(p.Path == "/job/exec/list" && p.Method == "POST") {
			roles = append(roles, "9528")
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
