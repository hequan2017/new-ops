// Package initialize ziqiang 自强-理工科大模型插件初始化（表/API/菜单/授权/路由，全部幂等）
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

	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/router"
)

// Gorm 注册 ziqiang 数据表（幂等）
func Gorm(ctx context.Context) {
	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.ZqTextCorpus),
		new(model.ZqComparePair),
		new(model.ZqImageCorpus),
		new(model.ZqEntity),
		new(model.ZqEdge),
		new(model.ZqRegulation),
		new(model.ZqCase),
		new(model.ZqEvidence),
		new(model.ZqPromptTemplate),
	)
	if err != nil {
		err = errors.Wrap(err, "ziqiang 注册表失败!")
		global.GVA_LOG.Error(fmt.Sprintf("%+v", err))
	}
}

// Router 注册 ziqiang 路由（private 组显式挂 JWT+casbin）
func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("")
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.RouterGroupApp.Init(private)
}

// Api 注册 ziqiang API 记录
func Api(ctx context.Context) {
	entities := []sysModel.SysApi{
		{Path: "/ziqiang/ping", Description: "自强理工科大模型连通性探测", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/textCorpus/list", Description: "文本语料列表", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/textCorpus/import", Description: "批量导入文本语料", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/textCorpus", Description: "保存文本语料", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/textCorpus", Description: "删除文本语料", ApiGroup: "自强理工科大模型", Method: "DELETE"},
		{Path: "/ziqiang/comparePair/list", Description: "对比样本对列表", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/comparePair", Description: "保存对比样本对", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/imageCorpus/list", Description: "图像语料列表", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/imageCorpus/compare", Description: "aHash 比对", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/imageCorpus", Description: "登记图像语料", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/termCheck", Description: "法律术语校验", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/dpMask", Description: "数值差分隐私脱敏", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/legalTokenize", Description: "法律文本分词", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/plagiarismCheck", Description: "隐性抄袭识别", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/entity/list", Description: "实体列表", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/entity", Description: "保存实体", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/entity", Description: "删除实体", ApiGroup: "自强理工科大模型", Method: "DELETE"},
		{Path: "/ziqiang/edge", Description: "保存关系边", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/edge", Description: "删除关系边", ApiGroup: "自强理工科大模型", Method: "DELETE"},
		{Path: "/ziqiang/graph", Description: "图谱全图", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/graph/bfs", Description: "图谱多跳推理", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/regulation/list", Description: "法规列表", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/regulation/search", Description: "法规检索", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/regulation", Description: "保存法规", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/regulation", Description: "删除法规", ApiGroup: "自强理工科大模型", Method: "DELETE"},
		{Path: "/ziqiang/case/list", Description: "判例列表", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/case/search", Description: "判例检索", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/case", Description: "保存判例", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/evidence/list", Description: "存证列表", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/evidence/verify", Description: "链完整性校验", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/evidence/head", Description: "链头摘要", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/evidence", Description: "存证详情", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/evidence", Description: "追加存证", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/prompt/list", Description: "提示词模板列表", ApiGroup: "自强理工科大模型", Method: "GET"},
		{Path: "/ziqiang/prompt/render", Description: "渲染提示词", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/prompt", Description: "保存提示词模板", ApiGroup: "自强理工科大模型", Method: "POST"},
		{Path: "/ziqiang/prompt", Description: "删除提示词模板", ApiGroup: "自强理工科大模型", Method: "DELETE"},
		{Path: "/ziqiang/qa/patent", Description: "知识产权问答", ApiGroup: "自强理工科大模型", Method: "POST"},
	}
	pluginUtils.RegisterApis(entities...)
}

// Menu 注册 ziqiang 菜单（自强理工科大模型父级 + 自强工作台子页）
func Menu(ctx context.Context) {
	entities := []sysModel.SysBaseMenu{
		{
			Path:      "ziqiang",
			Name:      "ziqiang",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      8,
			Meta:      sysModel.Meta{Title: "自强理工科大模型", Icon: "cpu"},
		},
		{
			Path:      "ziqiangLab",
			Name:      "ziqiangLab",
			Hidden:    false,
			Component: "plugin/ziqiang/view/lab/index.vue",
			Sort:      1,
			Meta:      sysModel.Meta{Title: "自强工作台", Icon: "search"},
		},
	}
	pluginUtils.RegisterMenus(entities...)
}

// Casbin 注册 ziqiang 策略（幂等）：仅 888 全量
func Casbin(ctx context.Context) {
	e := srvUtils.GetCasbin()
	if e == nil {
		return
	}
	for _, p := range []struct{ Path, Method string }{
		{"/ziqiang/ping", "GET"},
		{"/ziqiang/textCorpus/list", "GET"},
		{"/ziqiang/textCorpus/import", "POST"},
		{"/ziqiang/textCorpus", "POST"},
		{"/ziqiang/textCorpus", "DELETE"},
		{"/ziqiang/comparePair/list", "GET"},
		{"/ziqiang/comparePair", "POST"},
		{"/ziqiang/imageCorpus/list", "GET"},
		{"/ziqiang/imageCorpus/compare", "GET"},
		{"/ziqiang/imageCorpus", "POST"},
		{"/ziqiang/termCheck", "POST"},
		{"/ziqiang/dpMask", "POST"},
		{"/ziqiang/legalTokenize", "POST"},
		{"/ziqiang/plagiarismCheck", "POST"},
		{"/ziqiang/entity/list", "GET"},
		{"/ziqiang/entity", "POST"},
		{"/ziqiang/entity", "DELETE"},
		{"/ziqiang/edge", "POST"},
		{"/ziqiang/edge", "DELETE"},
		{"/ziqiang/graph", "GET"},
		{"/ziqiang/graph/bfs", "GET"},
		{"/ziqiang/regulation/list", "GET"},
		{"/ziqiang/regulation/search", "GET"},
		{"/ziqiang/regulation", "POST"},
		{"/ziqiang/regulation", "DELETE"},
		{"/ziqiang/case/list", "GET"},
		{"/ziqiang/case/search", "GET"},
		{"/ziqiang/case", "POST"},
		{"/ziqiang/evidence/list", "GET"},
		{"/ziqiang/evidence/verify", "GET"},
		{"/ziqiang/evidence/head", "GET"},
		{"/ziqiang/evidence", "GET"},
		{"/ziqiang/evidence", "POST"},
		{"/ziqiang/prompt/list", "GET"},
		{"/ziqiang/prompt/render", "POST"},
		{"/ziqiang/prompt", "POST"},
		{"/ziqiang/prompt", "DELETE"},
		{"/ziqiang/qa/patent", "POST"},
	} {
		has, err := e.HasPolicy("888", p.Path, p.Method)
		if err != nil || has {
			continue
		}
		if _, err := e.AddPolicy("888", p.Path, p.Method); err != nil {
			global.GVA_LOG.Error("ziqiang casbin 添加失败: " + err.Error())
		}
	}
}
