// Package service asset 插件业务逻辑
package service

import (
	"fmt"
	"net"

	"github.com/hequan2017/new-ops/server/plugin/asset/model"
)

// asset 插件错误码段：1000-1099（DEV_PLAN 3.6）
const (
	ErrCodeHostnameRequired = 1001
	ErrCodeIPInvalid        = 1002
	ErrCodeIPDuplicate      = 1003
	ErrCodeStatusInvalid    = 1004
)

// ServiceError 带业务错误码的错误（前端可按码分支处理）
type ServiceError struct {
	Code int
	Msg  string
}

func (e *ServiceError) Error() string { return e.Msg }

func newServiceErr(code int, msg string) *ServiceError {
	return &ServiceError{Code: code, Msg: msg}
}

// validateHost 主机资产校验（纯逻辑，不依赖 DB，可单测）
func validateHost(h *model.AssetHost) error {
	if h == nil {
		return newServiceErr(ErrCodeHostnameRequired, "主机信息不能为空")
	}
	if h.Hostname == "" {
		return newServiceErr(ErrCodeHostnameRequired, "主机名不能为空")
	}
	if net.ParseIP(h.IP) == nil {
		return newServiceErr(ErrCodeIPInvalid, fmt.Sprintf("内网IP格式不合法: %s", h.IP))
	}
	if h.PublicIP != "" && net.ParseIP(h.PublicIP) == nil {
		return newServiceErr(ErrCodeIPInvalid, fmt.Sprintf("公网IP格式不合法: %s", h.PublicIP))
	}
	switch h.Status {
	case "":
		h.Status = model.AssetStatusRunning // 默认状态
	case model.AssetStatusRunning, model.AssetStatusStopped,
		model.AssetStatusMaintain, model.AssetStatusRetired:
	default:
		return newServiceErr(ErrCodeStatusInvalid, fmt.Sprintf("非法的资产状态: %s", h.Status))
	}
	return nil
}
