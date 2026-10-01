// Package initialize asset 插件菜单种子（幂等）
package initialize

import (
	"context"

	model "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
)

// Menu 注册资产中心菜单（本场仅主机管理页；机房/产品线菜单随页面上线时追加）
func Menu(ctx context.Context) {
	entities := []model.SysBaseMenu{
		{
			Path:      "asset",
			Name:      "asset",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      2,
			Meta:      model.Meta{Title: "资产中心", Icon: "boxes"},
		},
		{
			Path:      "assetHost",
			Name:      "assetHost",
			Hidden:    false,
			Component: "plugin/asset/view/host/index.vue",
			Sort:      1,
			Meta:      model.Meta{Title: "主机管理", Icon: "monitor"},
		},
	}
	utils.RegisterMenus(entities...)
}
