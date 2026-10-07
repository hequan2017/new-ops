// Package initialize asset 插件菜单种子（幂等）
package initialize

import (
	"context"

	model "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
)

// Menu 注册资产中心菜单：资产中心（父）→ 主机管理/机房与机柜/产品线
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
		{
			Path:      "assetRoom",
			Name:      "assetRoom",
			Hidden:    false,
			Component: "plugin/asset/view/room/index.vue",
			Sort:      2,
			Meta:      model.Meta{Title: "机房与机柜", Icon: "office-building"},
		},
		{
			Path:      "assetProductLine",
			Name:      "assetProductLine",
			Hidden:    false,
			Component: "plugin/asset/view/productLine/index.vue",
			Sort:      3,
			Meta:      model.Meta{Title: "产品线", Icon: "collection"},
		},
		{
			Path:      "assetGroup",
			Name:      "assetGroup",
			Hidden:    false,
			Component: "plugin/asset/view/group/index.vue",
			Sort:      4,
			Meta:      model.Meta{Title: "资产组与权限", Icon: "key"},
		},
		{
			Path:      "assetCredential",
			Name:      "assetCredential",
			Hidden:    false,
			Component: "plugin/asset/view/credential/index.vue",
			Sort:      5,
			Meta:      model.Meta{Title: "凭据保险库", Icon: "lock"},
		},
	}
	utils.RegisterMenus(entities...)
}
