// Package initialize artsci 学科工具集插件初始化（表/API/菜单/授权/路由，全部幂等）
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

	"github.com/hequan2017/new-ops/server/plugin/artsci/model"
	"github.com/hequan2017/new-ops/server/plugin/artsci/router"
)

// Gorm 注册 artsci 数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.ArtsciTool),
		new(model.ArtsciRunLog),
		new(model.ArtsciGrading),
		new(model.ArtsciQuestion),
		new(model.ArtsciPaper),
		new(model.ArtsciSceneTemplate),
		new(model.ArtsciImageTask),
		new(model.ArtsciImageAsset),
		new(model.ArtsciInterpretation),
	)
	if err != nil {
		err = errors.Wrap(err, "artsci 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册 artsci 路由（private 组显式挂 JWT+casbin）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}

// Api 注册 artsci API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/artsci/ping", Description: "学科工具集连通性探测", ApiGroup: "学科工具集", Method: "GET"},
		{Path: "/artsci/tool/list", Description: "学科工具列表", ApiGroup: "学科工具集", Method: "GET"},
		{Path: "/artsci/tool", Description: "保存学科工具", ApiGroup: "学科工具集", Method: "POST"},
		{Path: "/artsci/tool/run", Description: "执行学科工具", ApiGroup: "学科工具集", Method: "POST"},
		{Path: "/artsci/tool/runLogs", Description: "工具执行留痕", ApiGroup: "学科工具集", Method: "GET"},
		{Path: "/artsci/tool", Description: "删除学科工具", ApiGroup: "学科工具集", Method: "DELETE"},
		{Path: "/artsci/grading", Description: "提交作业批改", ApiGroup: "学科工具集", Method: "POST"},
		{Path: "/artsci/grading/list", Description: "批改记录列表", ApiGroup: "学科工具集", Method: "GET"},
		{Path: "/artsci/question/generate", Description: "自动出题", ApiGroup: "学科工具集", Method: "POST"},
		{Path: "/artsci/question/list", Description: "题目列表", ApiGroup: "学科工具集", Method: "GET"},
		{Path: "/artsci/paper", Description: "规则组卷", ApiGroup: "学科工具集", Method: "POST"},
		{Path: "/artsci/paper/list", Description: "试卷列表", ApiGroup: "学科工具集", Method: "GET"},
		{Path: "/artsci/scene/list", Description: "文生图场景模板列表", ApiGroup: "学科工具集", Method: "GET"},
		{Path: "/artsci/imageTask", Description: "提交文生图任务", ApiGroup: "学科工具集", Method: "POST"},
		{Path: "/artsci/imageTask/list", Description: "文生图任务列表", ApiGroup: "学科工具集", Method: "GET"},
		{Path: "/artsci/imageAsset", Description: "登记图像资产", ApiGroup: "学科工具集", Method: "POST"},
		{Path: "/artsci/imageAsset/list", Description: "图像资产列表", ApiGroup: "学科工具集", Method: "GET"},
		{Path: "/artsci/imageAsset/interpret", Description: "图像解读与分析", ApiGroup: "学科工具集", Method: "POST"},
		{Path: "/artsci/interpretation/list", Description: "解读记录列表", ApiGroup: "学科工具集", Method: "GET"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 artsci 菜单（首元素为学科工具集父级，子页面挂其下）
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "artsci",
			Name:      "artsci",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      6,
			Meta:      sysModel.Meta{Title: "学科工具集", Icon: "collection"},
		},
		{
			Path:      "artsciLab",
			Name:      "artsciLab",
			Hidden:    false,
			Component: "plugin/artsci/view/lab/index.vue",
			Sort:      1,
			Meta:      sysModel.Meta{Title: "学科工具台", Icon: "edit-pen"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册 artsci 策略（幂等）：仅 888 全量
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/artsci/ping", "GET"},
		{"/artsci/tool/list", "GET"},
		{"/artsci/tool", "POST"},
		{"/artsci/tool/run", "POST"},
		{"/artsci/tool/runLogs", "GET"},
		{"/artsci/tool", "DELETE"},
		{"/artsci/grading", "POST"},
		{"/artsci/grading/list", "GET"},
		{"/artsci/question/generate", "POST"},
		{"/artsci/question/list", "GET"},
		{"/artsci/paper", "POST"},
		{"/artsci/paper/list", "GET"},
		{"/artsci/scene/list", "GET"},
		{"/artsci/imageTask", "POST"},
		{"/artsci/imageTask/list", "GET"},
		{"/artsci/imageAsset", "POST"},
		{"/artsci/imageAsset/list", "GET"},
		{"/artsci/imageAsset/interpret", "POST"},
		{"/artsci/interpretation/list", "GET"},
	} {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			global.GVA_LOG.Error("artsci casbin 添加失败: " + err.Error())
		}
	}
}
