// Package initialize fanhua 繁花-文科大模型插件初始化（表/API/菜单/授权/路由，全部幂等）
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

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/router"
)

// Gorm 注册 fanhua 数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.FanhuaCorpus),
		new(model.FanhuaParallel),
		new(model.FanhuaEntity),
		new(model.FanhuaEntityHit),
		new(model.FanhuaTextLabel),
		new(model.FanhuaImageAsset),
		new(model.FanhuaImageLabel),
		new(model.FanhuaVoiceEntry),
		new(model.FanhuaSensitiveWord),
		new(model.FanhuaComplianceTemplate),
		new(model.FanhuaConflictCase),
		new(model.FanhuaGovIndicator),
		new(model.FanhuaGovDataPoint),
		new(model.FanhuaPromptTemplate),
	)
	if err != nil {
		err = errors.Wrap(err, "fanhua 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册 fanhua 路由（private 组显式挂 JWT+casbin）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}

// Api 注册 fanhua API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/fanhua/ping", Description: "繁花文科大模型连通性探测", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/corpus/list", Description: "沪语语料列表", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/corpus", Description: "保存沪语语料", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/corpus/import", Description: "批量导入语料", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/corpus", Description: "删除沪语语料", ApiGroup: "繁花文科大模型", Method: "DELETE"},
		{Path: "/fanhua/parallel/list", Description: "平行语料列表", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/parallel/search", Description: "普通话对照检索", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/parallel", Description: "保存平行语料", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/parallel", Description: "删除平行语料", ApiGroup: "繁花文科大模型", Method: "DELETE"},
		{Path: "/fanhua/clean/dry", Description: "清洗试运行", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/clean", Description: "语料批量清洗", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/tokenize", Description: "沪语分词", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/entity/list", Description: "实体列表", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/entity/hits", Description: "实体标注记录", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/entity/annotate", Description: "实体标注", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/entity", Description: "保存实体", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/entity", Description: "删除实体", ApiGroup: "繁花文科大模型", Method: "DELETE"},
		{Path: "/fanhua/label/list", Description: "标注列表", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/label/stats", Description: "标注分布统计", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/label/build", Description: "构建标注", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/label", Description: "人工改标", ApiGroup: "繁花文科大模型", Method: "PUT"},
		{Path: "/fanhua/imageAsset/list", Description: "图片资产列表", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/imageAsset", Description: "登记图片资产", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/imageAsset/label", Description: "图片标注", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/imageLabel/list", Description: "图片标签列表", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/voice/list", Description: "语音条目列表", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/voice/search", Description: "海派文化组合检索", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/voice", Description: "保存语音条目", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/voice", Description: "删除语音条目", ApiGroup: "繁花文科大模型", Method: "DELETE"},
		{Path: "/fanhua/word/list", Description: "敏感词列表", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/word", Description: "保存敏感词", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/word", Description: "删除敏感词", ApiGroup: "繁花文科大模型", Method: "DELETE"},
		{Path: "/fanhua/values/health", Description: "价值观健康度评估", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/template/list", Description: "合规模板列表", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/conflict/list", Description: "冲突案例列表", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/govIndicator/list", Description: "治理指标树", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/govIndicator", Description: "保存治理指标", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/govIndicator", Description: "删除治理指标", ApiGroup: "繁花文科大模型", Method: "DELETE"},
		{Path: "/fanhua/govData", Description: "写入指标数据点", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/govData/list", Description: "指标数据点列表", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/govProfile", Description: "社区画像", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/prompt/list", Description: "提示词模板列表", ApiGroup: "繁花文科大模型", Method: "GET"},
		{Path: "/fanhua/prompt/render", Description: "提示词渲染", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/prompt", Description: "保存提示词模板", ApiGroup: "繁花文科大模型", Method: "POST"},
		{Path: "/fanhua/prompt", Description: "删除提示词模板", ApiGroup: "繁花文科大模型", Method: "DELETE"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 fanhua 菜单（繁花文科大模型父级，子页面随 N16 挂其下）
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "fanhua",
			Name:      "fanhua",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      7,
			Meta:      sysModel.Meta{Title: "繁花文科大模型", Icon: "notebook-2"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册 fanhua 策略（幂等）：仅 888 全量
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/fanhua/ping", "GET"},
		{"/fanhua/corpus/list", "GET"},
		{"/fanhua/corpus", "POST"},
		{"/fanhua/corpus/import", "POST"},
		{"/fanhua/corpus", "DELETE"},
		{"/fanhua/parallel/list", "GET"},
		{"/fanhua/parallel/search", "GET"},
		{"/fanhua/parallel", "POST"},
		{"/fanhua/parallel", "DELETE"},
		{"/fanhua/clean/dry", "POST"},
		{"/fanhua/clean", "POST"},
		{"/fanhua/tokenize", "POST"},
		{"/fanhua/entity/list", "GET"},
		{"/fanhua/entity/hits", "GET"},
		{"/fanhua/entity/annotate", "POST"},
		{"/fanhua/entity", "POST"},
		{"/fanhua/entity", "DELETE"},
		{"/fanhua/label/list", "GET"},
		{"/fanhua/label/stats", "GET"},
		{"/fanhua/label/build", "POST"},
		{"/fanhua/label", "PUT"},
		{"/fanhua/imageAsset/list", "GET"},
		{"/fanhua/imageAsset", "POST"},
		{"/fanhua/imageAsset/label", "POST"},
		{"/fanhua/imageLabel/list", "GET"},
		{"/fanhua/voice/list", "GET"},
		{"/fanhua/voice/search", "POST"},
		{"/fanhua/voice", "POST"},
		{"/fanhua/voice", "DELETE"},
		{"/fanhua/word/list", "GET"},
		{"/fanhua/word", "POST"},
		{"/fanhua/word", "DELETE"},
		{"/fanhua/values/health", "POST"},
		{"/fanhua/template/list", "GET"},
		{"/fanhua/conflict/list", "GET"},
		{"/fanhua/govIndicator/list", "GET"},
		{"/fanhua/govIndicator", "POST"},
		{"/fanhua/govIndicator", "DELETE"},
		{"/fanhua/govData", "POST"},
		{"/fanhua/govData/list", "GET"},
		{"/fanhua/govProfile", "GET"},
		{"/fanhua/prompt/list", "GET"},
		{"/fanhua/prompt/render", "POST"},
		{"/fanhua/prompt", "POST"},
		{"/fanhua/prompt", "DELETE"},
	} {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			global.GVA_LOG.Error("fanhua casbin 添加失败: " + err.Error())
		}
	}
}
