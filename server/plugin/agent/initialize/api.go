// Package initialize 白泽 Agent 通道初始化
package initialize

import (
	"context"

	sysModel "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
	serverUtils "github.com/hequan2017/new-ops/server/utils"
	"go.uber.org/zap"
)

// Api 注册 Agent API 记录
func Api(ctx context.Context) {
	apis := []sysModel.SysApi{
		{Path: "/agent/instance/token", Description: "签发/重置 Agent 接入令牌", ApiGroup: "Agent通道", Method: "POST"},
		{Path: "/agent/instance/list", Description: "Agent 实例列表", ApiGroup: "Agent通道", Method: "GET"},
		{Path: "/agent/instance", Description: "吊销 Agent 注册", ApiGroup: "Agent通道", Method: "DELETE"},
	}
	utils.RegisterApis(apis...)
}

// Casbin 注册 Agent 策略（幂等）：令牌签发与吊销仅 888；实例列表 888+9528 只读
func Casbin(ctx context.Context) {
	e := serverUtils.GetCasbin()
	if e == nil {
		zap.L().Warn("agent 插件：casbin 未初始化，跳过策略注册")
		return
	}
	all := []struct {
		Path   string
		Method string
	}{
		{"/agent/instance/token", "POST"},
		{"/agent/instance/list", "GET"},
		{"/agent/instance", "DELETE"},
	}
	readonly := []struct {
		Path   string
		Method string
	}{
		{"/agent/instance/list", "GET"},
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
				zap.L().Error("agent 插件：添加 casbin 策略失败", zap.Error(err), zap.String("role", role))
			}
		}
	}
	apply("888", all)
	apply("9528", readonly)
}
