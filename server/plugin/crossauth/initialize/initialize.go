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
		new(model.CaRoleTemplate),
		new(model.CaAutoGrantRule),
		new(model.CaGrantRequest),
		new(model.CaGrantAssignment),
		new(model.CaPermGrant),
		new(model.CaGrantExpiry),
		new(model.CaExpiryReminder),
		new(model.CaDataClassify),
		new(model.CaKeyRotation),
		new(model.CaAccessLog),
		new(model.CaAlert),
		new(model.CaComputeUsage),
		new(model.CaComputeAlert),
		new(model.CaAuditEvent),
		new(model.CaOpsTicket),
		new(model.CaOpsMessage),
		new(model.CaRemoteAssist),
		new(model.CaPlaybook),
		new(model.CaDrillPlan),
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
		{Path: "/crossauth/roleTemplate/list", Description: "角色模板列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/roleTemplate", Description: "保存角色模板", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/grant/rule", Description: "保存自动匹配规则", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/grant/auto", Description: "自动匹配分配", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/grant/request", Description: "提交审核申请", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/grant/approve", Description: "审核推进", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/grant/reject", Description: "申请驳回", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/grant/requests", Description: "申请列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/grant/assignments", Description: "分配列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/matrix/list", Description: "四级授权列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/matrix/check", Description: "四级权限校验", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/matrix", Description: "保存四级授权", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/matrix", Description: "删除四级授权", ApiGroup: "跨校认证与安全", Method: "DELETE"},
		{Path: "/crossauth/lifecycle/list", Description: "有效期列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/lifecycle/reminders", Description: "到期提醒列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/lifecycle", Description: "登记有效期", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/lifecycle/scan", Description: "到期扫描", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/classify/scan", Description: "分级批量扫描", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/classify/list", Description: "分级结果列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/sm4/keys", Description: "密钥轮换表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/sm4/rotate", Description: "密钥轮换", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/sm4/encrypt", Description: "SM4 加密", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/sm4/decrypt", Description: "SM4 解密", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/fieldCrypt/encrypt", Description: "字段加密", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/fieldCrypt/decrypt", Description: "字段解密", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/fieldCrypt/watermark", Description: "访问水印", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/access/check", Description: "资源访问校验", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/access/logs", Description: "访问日志", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/access/alerts", Description: "阻断告警列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/access/handle", Description: "告警处置", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/compute/usage", Description: "记录算力用量", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/compute/usage/list", Description: "算力用量列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/compute/alerts", Description: "算力告警列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/compute/handle", Description: "算力告警处置", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/monitor/fiveDim", Description: "五维监控看板", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/auditEvent/list", Description: "三级审计查询", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/auditEvent", Description: "登记审计事件", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/opsTicket/list", Description: "运维工单列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/opsTicket/advance", Description: "工单推进", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/opsTicket/message", Description: "协同消息", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/opsTicket/messages", Description: "消息列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/opsTicket", Description: "创建运维工单", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/remoteAssist", Description: "远程协助授权", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/remoteAssist/revoke", Description: "撤销授权", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/remoteAssist/list", Description: "授权记录", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/playbook/list", Description: "应急预案列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/playbook", Description: "保存应急预案", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/drill/list", Description: "演练计划列表", ApiGroup: "跨校认证与安全", Method: "GET"},
		{Path: "/crossauth/drill/finish", Description: "演练复盘归档", ApiGroup: "跨校认证与安全", Method: "POST"},
		{Path: "/crossauth/drill", Description: "新建演练计划", ApiGroup: "跨校认证与安全", Method: "POST"},
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
		{"/crossauth/roleTemplate/list", "GET"},
		{"/crossauth/roleTemplate", "POST"},
		{"/crossauth/grant/rule", "POST"},
		{"/crossauth/grant/auto", "POST"},
		{"/crossauth/grant/request", "POST"},
		{"/crossauth/grant/approve", "POST"},
		{"/crossauth/grant/reject", "POST"},
		{"/crossauth/grant/requests", "GET"},
		{"/crossauth/grant/assignments", "GET"},
		{"/crossauth/matrix/list", "GET"},
		{"/crossauth/matrix/check", "POST"},
		{"/crossauth/matrix", "POST"},
		{"/crossauth/matrix", "DELETE"},
		{"/crossauth/lifecycle/list", "GET"},
		{"/crossauth/lifecycle/reminders", "GET"},
		{"/crossauth/lifecycle", "POST"},
		{"/crossauth/lifecycle/scan", "POST"},
		{"/crossauth/classify/scan", "POST"},
		{"/crossauth/classify/list", "GET"},
		{"/crossauth/sm4/keys", "GET"},
		{"/crossauth/sm4/rotate", "POST"},
		{"/crossauth/sm4/encrypt", "POST"},
		{"/crossauth/sm4/decrypt", "POST"},
		{"/crossauth/fieldCrypt/encrypt", "POST"},
		{"/crossauth/fieldCrypt/decrypt", "POST"},
		{"/crossauth/fieldCrypt/watermark", "POST"},
		{"/crossauth/access/check", "POST"},
		{"/crossauth/access/logs", "GET"},
		{"/crossauth/access/alerts", "GET"},
		{"/crossauth/access/handle", "POST"},
		{"/crossauth/compute/usage", "POST"},
		{"/crossauth/compute/usage/list", "GET"},
		{"/crossauth/compute/alerts", "GET"},
		{"/crossauth/compute/handle", "POST"},
		{"/crossauth/monitor/fiveDim", "GET"},
		{"/crossauth/auditEvent/list", "GET"},
		{"/crossauth/auditEvent", "POST"},
		{"/crossauth/opsTicket/list", "GET"},
		{"/crossauth/opsTicket/advance", "POST"},
		{"/crossauth/opsTicket/message", "POST"},
		{"/crossauth/opsTicket/messages", "GET"},
		{"/crossauth/opsTicket", "POST"},
		{"/crossauth/remoteAssist", "POST"},
		{"/crossauth/remoteAssist/revoke", "POST"},
		{"/crossauth/remoteAssist/list", "GET"},
		{"/crossauth/playbook/list", "GET"},
		{"/crossauth/playbook", "POST"},
		{"/crossauth/drill/list", "GET"},
		{"/crossauth/drill/finish", "POST"},
		{"/crossauth/drill", "POST"},
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
