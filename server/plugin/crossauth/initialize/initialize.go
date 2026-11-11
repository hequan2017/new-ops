// Package initialize crossauth 跨校认证与安全插件初始化（表/API/菜单/授权/路由，全部幂等）
package initialize

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/middleware"
	sysModel "github.com/hequan2017/new-ops/server/model/system"
	pluginUtils "github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
	srvUtils "github.com/hequan2017/new-ops/server/utils"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/router"
)

// Gorm 注册 crossauth 数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.CaFederation),
		new(model.CaSyncUser),
		new(model.CaSyncTask),
		new(model.CaTrustRule),
		new(model.CaMfaConfig),
		new(model.CaMfaPolicy),
		new(model.CaSsoApp),
		new(model.CaSsoTicket),
	)
	if err != nil {
		err = errors.Wrap(err, "crossauth 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册 crossauth 路由（private 组显式挂 JWT+casbin）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}

// Api 注册 crossauth API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/crossauth/ping", Description: "跨校认证与安全连通性探测", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/federation/list", Description: "高校对接列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/federation/tasks", Description: "同步任务记录", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/syncUser/list", Description: "同步用户列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/federation", Description: "保存对接配置", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/federation/sync", Description: "执行用户同步", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/trust/list", Description: "互认规则列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/trust/check", Description: "互认校验", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/trust", Description: "保存互认规则", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/trust", Description: "删除互认规则", ApiGroup: "跨校认证与安全", Method: "DELETE"},
		{Path: "/crossauth/mfa/list", Description: "MFA 配置列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/mfa/verify", Description: "MFA 校验", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/mfa/policy/list", Description: "MFA 策略列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/mfa", Description: "登记 MFA", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/mfa/policy", Description: "保存 MFA 策略", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/sso/apps", Description: "SSO 应用列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/sso/ticket", Description: "SSO 票据签发", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/sso/redeem", Description: "SSO 票据换取", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/sso/app", Description: "注册 SSO 应用", ApiGroup: "跨校认证与安全", Method: "POST"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 crossauth 菜单（跨校认证与安全父级，子页面随 Q19 挂其下）
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "crossauth",
			Name:      "crossauth",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      10,
			Meta:      sysModel.Meta{Title: "跨校认证与安全", Icon: "lock"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册 crossauth 策略（幂等）：仅 888 全量
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/crossauth/ping", "GET"},
		{"/crossauth/federation/list", "GET"},
		{"/crossauth/federation/tasks", "GET"},
		{"/crossauth/syncUser/list", "GET"},
		{"/crossauth/federation", "POST"},
		{"/crossauth/federation/sync", "POST"},
		{"/crossauth/trust/list", "GET"},
		{"/crossauth/trust/check", "POST"},
		{"/crossauth/trust", "POST"},
		{"/crossauth/trust", "DELETE"},
		{"/crossauth/mfa/list", "GET"},
		{"/crossauth/mfa/verify", "POST"},
		{"/crossauth/mfa/policy/list", "GET"},
		{"/crossauth/mfa", "POST"},
		{"/crossauth/mfa/policy", "POST"},
		{"/crossauth/sso/apps", "GET"},
		{"/crossauth/sso/ticket", "POST"},
		{"/crossauth/sso/redeem", "POST"},
		{"/crossauth/sso/app", "POST"},
	} {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			global.GVA_LOG.Error("crossauth casbin 添加失败: " + err.Error())
		}
	}
}
