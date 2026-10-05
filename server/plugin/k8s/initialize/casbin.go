// Package initialize k8s 插件授权种子（幂等）
package initialize

import (
	"github.com/hequan2017/new-ops/server/utils"
	"go.uber.org/zap"
)

// Casbin 注册 k8s 策略：888 全部；9528 只读（list/test）
func Casbin() {
	e := utils.GetCasbin()
	if e == nil {
		zap.L().Warn("k8s 插件：casbin 未初始化，跳过策略注册")
		return
	}
	all := []struct {
		Path   string
		Method string
	}{
		{"/k8s/cluster/create", "POST"},
		{"/k8s/cluster/delete", "DELETE"},
		{"/k8s/cluster/list", "GET"},
		{"/k8s/cluster/test", "GET"},
	}
	readonly := []struct {
		Path   string
		Method string
	}{
		{"/k8s/cluster/list", "GET"},
		{"/k8s/cluster/test", "GET"},
	}
	apply := func(role string, rules []struct {
		Path   string
		Method string
	}) {
		for _, p := range rules {
			has, err := e.HasPolicy(role, p.Path, p.Method)
			if err != nil || has {
				continue
			}
			if _, err := e.AddPolicy(role, p.Path, p.Method); err != nil {
				zap.L().Error("k8s 插件：添加 casbin 策略失败", zap.Error(err), zap.String("role", role))
			}
		}
	}
	apply("888", all)
	apply("9528", readonly)
}
