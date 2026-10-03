// Package initialize 白泽终端插件初始化
package initialize

import (
	"context"

	model "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
)

// Menu 注册终端菜单（独立终端页入口；主机管理页亦可发起）
func Menu(ctx context.Context) {
	entities := []model.SysBaseMenu{
		{
			Path:      "term",
			Name:      "term",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      3,
			Meta:      model.Meta{Title: "终端作业", Icon: "monitor"},
		},
		{
			Path:      "termTerminal",
			Name:      "termTerminal",
			Hidden:    false,
			Component: "plugin/term/view/terminal/index.vue",
			Sort:      1,
			Meta:      model.Meta{Title: "Web 终端", Icon: "platform"},
		},
	}
	utils.RegisterMenus(entities...)
}
