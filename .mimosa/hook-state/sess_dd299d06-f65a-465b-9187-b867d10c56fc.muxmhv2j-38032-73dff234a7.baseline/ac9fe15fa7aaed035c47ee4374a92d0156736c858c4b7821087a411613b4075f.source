// Package initialize 白泽流水线菜单种子
package initialize

import (
	"context"

	model "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
)

// Menu 注册流水线菜单
func Menu(ctx context.Context) {
	entities := []model.SysBaseMenu{
		{
			Path:      "pipeline",
			Name:      "pipeline",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      5,
			Meta:      model.Meta{Title: "流水线", Icon: "share"},
		},
		{
			Path:      "pipelineList",
			Name:      "pipelineList",
			Hidden:    false,
			Component: "plugin/pipeline/view/pipeline/index.vue",
			Sort:      1,
			Meta:      model.Meta{Title: "流水线管理", Icon: "set-up"},
		},
	}
	utils.RegisterMenus(entities...)
}
