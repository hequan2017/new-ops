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
		{"/k8s/cluster/overview", "GET"},
		{"/k8s/pod/list", "GET"},
		{"/k8s/pod/logs", "GET"},
		{"/k8s/pod/detail", "GET"},
		{"/k8s/deployment/list", "GET"},
		{"/k8s/statefulset/list", "GET"},
		{"/k8s/daemonset/list", "GET"},
		{"/k8s/workload/yaml", "GET"},
		{"/k8s/node/list", "GET"},
		{"/k8s/node/detail", "GET"},
		{"/k8s/service/list", "GET"},
		{"/k8s/configmap/list", "GET"},
		{"/k8s/secret/list", "GET"},
		{"/k8s/pvc/list", "GET"},
		{"/k8s/ingress/list", "GET"},
		{"/k8s/event/list", "GET"},
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
	// 写操作仅 888（扩缩容/滚动重启/节点隔离/节点驱逐/删除 Pod/YAML 下发）
	writeOps := []struct {
		Path   string
		Method string
	}{
		{"/k8s/deployment/scale", "POST"},
		{"/k8s/deployment/restart", "POST"},
		{"/k8s/node/cordon", "POST"},
		{"/k8s/node/drain", "POST"},
		{"/k8s/pod/delete", "POST"},
		{"/k8s/workload/diff", "POST"},
		{"/k8s/workload/apply", "POST"},
		// 终端为写级能力：execws 握手 casbin 自验（仅 888）
		{"/k8s/pod/execws", "GET"},
	}
	for _, p := range writeOps {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			zap.L().Error("k8s 插件：添加写操作策略失败", zap.Error(err))
		}
	}
}
