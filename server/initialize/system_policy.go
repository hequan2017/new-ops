// Package initialize 系统级 casbin 策略补丁（幂等，仅对存量库补缺）
// 背景：/autoCode/llmAuto 等 5 条路由原挂 public 组未鉴权，收敛至 private 后，
// 已初始化的库缺这批 888 策略（source 种子只在首次建库时写入），启动时补齐。
package initialize

import (
	"github.com/hequan2017/new-ops/server/utils"
	"go.uber.org/zap"
)

const systemPolicySuperAdmin = "888"

var systemPolicyPatches = []struct {
	Path   string
	Method string
}{
	{"/autoCode/llmAuto", "POST"},
	{"/autoCode/llmAutoSSE", "POST"},
	{"/autoCode/initMenu", "POST"},
	{"/autoCode/initAPI", "POST"},
	{"/autoCode/initDictionary", "POST"},
}

// EnsureSystemCasbinPolicies 为超管幂等补齐系统路由策略（已存在则跳过）
func EnsureSystemCasbinPolicies() {
	e := utils.GetCasbin()
	if e == nil {
		zap.L().Warn("系统策略补丁：casbin 未初始化，跳过")
		return
	}
	for _, p := range systemPolicyPatches {
		has, err := e.HasPolicy(systemPolicySuperAdmin, p.Path, p.Method)
		if err != nil {
			zap.L().Error("系统策略补丁：查询 casbin 策略失败", zap.Error(err), zap.String("path", p.Path))
			continue
		}
		if has {
			continue
		}
		if _, err := e.AddPolicy(systemPolicySuperAdmin, p.Path, p.Method); err != nil {
			zap.L().Error("系统策略补丁：添加 casbin 策略失败", zap.Error(err), zap.String("path", p.Path))
		}
	}
}
