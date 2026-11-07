// Package initialize exam 插件初始化（表/API/菜单/授权/路由，全部幂等）
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

	"github.com/hequan2017/new-ops/server/plugin/exam/model"
	"github.com/hequan2017/new-ops/server/plugin/exam/router"
)

// Gorm 注册 exam 数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.ExamQuestion),
		new(model.ExamPaper),
	)
	if err != nil {
		err = errors.Wrap(err, "exam 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册 exam 路由（private 组显式挂 JWT+casbin）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}

// Api 注册 exam API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/exam/question/list", Description: "考题列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/exam/paper/list", Description: "试卷列表", ApiGroup: "大模型平台", Method: "GET"},
		{Path: "/exam/paper", Description: "保存试卷", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/exam/paper/compose", Description: "试卷组卷", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/exam/paper", Description: "删除试卷", ApiGroup: "大模型平台", Method: "DELETE"},
		{Path: "/exam/question", Description: "保存考题", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/exam/question/import", Description: "考题 JSONL 导入", ApiGroup: "大模型平台", Method: "POST"},
		{Path: "/exam/question", Description: "删除考题", ApiGroup: "大模型平台", Method: "DELETE"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 exam 菜单（首元素为大模型平台父级引用）
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "llmops",
			Name:      "llmops",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      5,
			Meta:      sysModel.Meta{Title: "大模型平台", Icon: "data-line"},
		},
		{
			Path:      "examAdmin",
			Name:      "examAdmin",
			Hidden:    false,
			Component: "plugin/exam/view/admin/index.vue",
			Sort:      9,
			Meta:      sysModel.Meta{Title: "考试管理", Icon: "notebook-2"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册 exam 策略（幂等）：仅 888 全量
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/exam/question/list", "GET"},
		{"/exam/paper/list", "GET"},
		{"/exam/paper", "POST"},
		{"/exam/paper/compose", "POST"},
		{"/exam/paper", "DELETE"},
		{"/exam/question", "POST"},
		{"/exam/question/import", "POST"},
		{"/exam/question", "DELETE"},
	} {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			global.GVA_LOG.Error("exam casbin 添加失败: " + err.Error())
		}
	}
}
