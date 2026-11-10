// Package initialize eduadmin 教育管理一模通管插件初始化（表/API/菜单/授权/路由，全部幂等）
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

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/router"
)

// Gorm 注册 eduadmin 数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.EduRecord),
		new(model.EduPolicy),
		new(model.EduResume),
		new(model.EduPosition),
		new(model.EduPerfRecord),
		new(model.EduPerfReport),
		new(model.EduTalentStat),
		new(model.EduBudget),
		new(model.EduInvoice),
		new(model.EduRiskAlert),
		new(model.EduAsset),
		new(model.EduSensorReading),
		new(model.EduMaintPlan),
		new(model.EduPurchase),
		new(model.EduEnergyRecord),
	)
	if err != nil {
		err = errors.Wrap(err, "eduadmin 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册 eduadmin 路由（private 组显式挂 JWT+casbin）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}

// Api 注册 eduadmin API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/eduadmin/ping", Description: "教育管理一模通管连通性探测", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/record/list", Description: "数据记录列表", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/record/import", Description: "批量导入数据记录", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/record", Description: "保存数据记录", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/record", Description: "删除数据记录", ApiGroup: "教育管理一模通管", Method: "DELETE"},
		{Path: "/eduadmin/clean", Description: "数据批量清洗", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/mask", Description: "数据脱敏分级", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/mask/batch", Description: "批量脱敏预览", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/policy/list", Description: "知识库条目列表", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/policy/search", Description: "知识库检索", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/policy", Description: "保存知识库条目", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/policy", Description: "删除知识库条目", ApiGroup: "教育管理一模通管", Method: "DELETE"},
		{Path: "/eduadmin/resume/list", Description: "简历列表", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/resume", Description: "保存简历", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/position/list", Description: "岗位列表", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/position", Description: "保存岗位", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/screen", Description: "候选人筛选", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/perf/list", Description: "绩效数据列表", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/perf", Description: "保存绩效数据", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/perf/report", Description: "生成绩效报告", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/perf/report/list", Description: "绩效报告列表", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/talent/stat/list", Description: "年度人才统计列表", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/talent/stat", Description: "保存年度人才统计", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/talent/forecast", Description: "人才需求预测", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/budget/list", Description: "预算列表", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/budget/analysis", Description: "预算同比分析", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/budget/suggest", Description: "预算分配建议", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/budget", Description: "保存预算", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/budget/compare", Description: "预算方案对比", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/expense/list", Description: "报销列表", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/expense", Description: "提交报销", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/expense/pay", Description: "报销支付", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/expense/reject", Description: "报销驳回", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/risk/list", Description: "财务预警列表", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/risk/scan", Description: "财务风险扫描", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/risk/handle", Description: "预警标记处理", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/asset/list", Description: "设备列表", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/asset/maintenance", Description: "资产故障预测", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/asset/maintenance/list", Description: "维护计划列表", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/asset", Description: "保存设备", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/asset/reading", Description: "新增传感器读数", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/purchase/list", Description: "采购记录列表", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/purchase/suggest", Description: "智能采购建议", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/purchase", Description: "保存采购记录", ApiGroup: "教育管理一模通管", Method: "POST"},
		{Path: "/eduadmin/energy/list", Description: "能耗记录列表", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/energy/analyze", Description: "能耗聚合分析", ApiGroup: "教育管理一模通管", Method: "GET"},
		{Path: "/eduadmin/energy", Description: "保存能耗记录", ApiGroup: "教育管理一模通管", Method: "POST"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 eduadmin 菜单（教育管理一模通管父级，子页面随 P31/P32 挂其下）
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "eduadmin",
			Name:      "eduadmin",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      9,
			Meta:      sysModel.Meta{Title: "教育管理一模通管", Icon: "data-board"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册 eduadmin 策略（幂等）：仅 888 全量
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/eduadmin/ping", "GET"},
		{"/eduadmin/record/list", "GET"},
		{"/eduadmin/record/import", "POST"},
		{"/eduadmin/record", "POST"},
		{"/eduadmin/record", "DELETE"},
		{"/eduadmin/clean", "POST"},
		{"/eduadmin/mask", "POST"},
		{"/eduadmin/mask/batch", "GET"},
		{"/eduadmin/policy/list", "GET"},
		{"/eduadmin/policy/search", "GET"},
		{"/eduadmin/policy", "POST"},
		{"/eduadmin/policy", "DELETE"},
		{"/eduadmin/resume/list", "GET"},
		{"/eduadmin/resume", "POST"},
		{"/eduadmin/position/list", "GET"},
		{"/eduadmin/position", "POST"},
		{"/eduadmin/screen", "GET"},
		{"/eduadmin/perf/list", "GET"},
		{"/eduadmin/perf", "POST"},
		{"/eduadmin/perf/report", "POST"},
		{"/eduadmin/perf/report/list", "GET"},
		{"/eduadmin/talent/stat/list", "GET"},
		{"/eduadmin/talent/stat", "POST"},
		{"/eduadmin/talent/forecast", "GET"},
		{"/eduadmin/budget/list", "GET"},
		{"/eduadmin/budget/analysis", "GET"},
		{"/eduadmin/budget/suggest", "GET"},
		{"/eduadmin/budget", "POST"},
		{"/eduadmin/budget/compare", "POST"},
		{"/eduadmin/expense/list", "GET"},
		{"/eduadmin/expense", "POST"},
		{"/eduadmin/expense/pay", "POST"},
		{"/eduadmin/expense/reject", "POST"},
		{"/eduadmin/risk/list", "GET"},
		{"/eduadmin/risk/scan", "POST"},
		{"/eduadmin/risk/handle", "POST"},
		{"/eduadmin/asset/list", "GET"},
		{"/eduadmin/asset/maintenance", "GET"},
		{"/eduadmin/asset/maintenance/list", "GET"},
		{"/eduadmin/asset", "POST"},
		{"/eduadmin/asset/reading", "POST"},
		{"/eduadmin/purchase/list", "GET"},
		{"/eduadmin/purchase/suggest", "GET"},
		{"/eduadmin/purchase", "POST"},
		{"/eduadmin/energy/list", "GET"},
		{"/eduadmin/energy/analyze", "GET"},
		{"/eduadmin/energy", "POST"},
	} {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			global.GVA_LOG.Error("eduadmin casbin 添加失败: " + err.Error())
		}
	}
}
