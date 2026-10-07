// Package initialize 白泽终端插件授权种子
package initialize

import (
	"context"

	"github.com/hequan2017/new-ops/server/utils"
	"go.uber.org/zap"
)

// Casbin 注册 WS 端点策略（幂等；握手时 casbin 自验强制生效）
// 终端/日志 tail 为写级能力（建立 SSH 通道），仅 888——历史误授 9528 的策略幂等移除
func Casbin(ctx context.Context) {
	e := utils.GetCasbin()
	if e == nil {
		zap.L().Warn("term 插件：casbin 未初始化，跳过策略注册")
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/term/ws", "GET"},
		{"/term/logtail", "GET"},
	} {
		// 历史遗留：9528 曾被授予终端权限，收敛移除
		if has, _ := e.HasPolicy("9528", p.Path, p.Method); has {
			if _, err := e.RemovePolicy("9528", p.Path, p.Method); err != nil {
				zap.L().Error("term 插件：移除 9528 终端策略失败", zap.Error(err))
			}
		}
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			zap.L().Error("term 插件：添加 casbin 策略失败", zap.Error(err))
		}
	}
	// SFTP：888 全部；9528 只读（list/download）
	for _, sp := range []struct {
		Path   string
		Method string
	}{
		{"/term/sftp/list", "GET"},
		{"/term/sftp/mkdir", "POST"},
		{"/term/sftp/delete", "POST"},
		{"/term/sftp/rename", "POST"},
		{"/term/sftp/upload", "POST"},
		{"/term/sftp/download", "GET"},
	} {
		roles := []string{"888"}
		if sp.Method == "GET" {
			roles = append(roles, "9528")
		}
		for _, role := range roles {
			has, err := e.HasPolicy(role, sp.Path, sp.Method)
			if err != nil || has {
				continue
			}
			if _, err := e.AddPolicy(role, sp.Path, sp.Method); err != nil {
				zap.L().Error("term 插件：添加 SFTP 策略失败", zap.Error(err))
			}
		}
	}
	// 审计查询仅 888
	for _, p := range []struct {
		Path   string
		Method string
	}{
		{"/term/session/list", "POST"},
		{"/term/session/streams", "GET"},
		{"/term/session/commands", "GET"},
	} {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			zap.L().Error("term 插件：添加审计策略失败", zap.Error(err))
		}
	}
}
