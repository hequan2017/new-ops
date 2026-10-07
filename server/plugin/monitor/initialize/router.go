// Package initialize 白泽监控插件初始化
package initialize

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/global"
	sysModel "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/middleware"
	"github.com/hequan2017/new-ops/server/plugin/monitor/model"
	"github.com/hequan2017/new-ops/server/plugin/monitor/router"
	"github.com/hequan2017/new-ops/server/plugin/monitor/service"
	pluginUtils "github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
	srvUtils "github.com/hequan2017/new-ops/server/utils"
	"github.com/pkg/errors"
)

// Gorm 注册监控数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.MonitorMetric),
		new(model.MonitorAlertRule),
		new(model.MonitorAlertEvent),
	)
	if err != nil {
		err = errors.Wrap(err, "monitor 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册监控路由（private 组显式挂 JWT+casbin——插件自建组不继承底座中间件）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}

// Api 注册监控 API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/monitor/metric/list", Description: "主机指标序列查询", ApiGroup: "监控告警", Method: "GET"},
		{Path: "/monitor/metric/collect", Description: "手动触发全量采集", ApiGroup: "监控告警", Method: "POST"},
		{Path: "/monitor/alert/rule", Description: "创建告警规则", ApiGroup: "监控告警", Method: "POST"},
		{Path: "/monitor/alert/rule", Description: "更新告警规则", ApiGroup: "监控告警", Method: "PUT"},
		{Path: "/monitor/alert/rule", Description: "删除告警规则", ApiGroup: "监控告警", Method: "DELETE"},
		{Path: "/monitor/alert/rule/list", Description: "告警规则列表", ApiGroup: "监控告警", Method: "GET"},
		{Path: "/monitor/alert/event/list", Description: "告警事件列表", ApiGroup: "监控告警", Method: "POST"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册监控菜单（图表入口在主机页「监控」抽屉）
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "monitor",
			Name:      "monitor",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      7,
			Meta:      sysModel.Meta{Title: "监控告警", Icon: "odometer"},
		},
		{
			Path:      "monitorAlert",
			Name:      "monitorAlert",
			Hidden:    false,
			Component: "plugin/monitor/view/alert/index.vue",
			Sort:      1,
			Meta:      sysModel.Meta{Title: "告警规则", Icon: "bell"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册监控策略（幂等）：888 全量；9528 只读查询
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/monitor/metric/list", "GET"},
		{"/monitor/metric/collect", "POST"},
		{"/monitor/alert/rule", "POST"},
		{"/monitor/alert/rule", "PUT"},
		{"/monitor/alert/rule", "DELETE"},
		{"/monitor/alert/rule/list", "GET"},
		{"/monitor/alert/event/list", "POST"},
	} {
		roles := []string{"888"}
		if p.Method == "GET" {
			roles = append(roles, "9528")
		}
		for _, role := range roles {
			has, err := e.HasPolicy(role, p.Path, p.Method)
			if err != nil || has {
				continue
			}
			if _, err := e.AddPolicy(role, p.Path, p.Method); err != nil {
				global.GVA_LOG.Error("monitor casbin 添加失败: " + err.Error())
			}
		}
	}
}

// Timer 注册定时采集（@every 5m，指标留存 30 天随轮清理）
func Timer(ctx context.Context) {
	svc := new(service.MonitorService)
	if _, err := global.GVA_Timer.AddTaskByFunc("monitor", "@every 5m", func() {
		svc.CollectAll()
	}, "monitor-perf-collect"); err != nil {
		global.GVA_LOG.Error("monitor 定时采集注册失败: " + err.Error())
	}
}
