// Package initialize k8s 插件菜单种子
package initialize

import (
	"context"

	model "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
)

// Menu 注册 K8s 管理菜单：K8s 管理（父）→ 集群管理
func Menu(ctx context.Context) {
	entities := []model.SysBaseMenu{
		{
			Path:      "k8s",
			Name:      "k8sMgmt",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      4,
			Meta:      model.Meta{Title: "K8s 管理", Icon: "kubernetes"},
		},
		{
			Path:      "k8sCluster",
			Name:      "k8sCluster",
			Hidden:    false,
			Component: "plugin/k8s/view/cluster/index.vue",
			Sort:      1,
			Meta:      model.Meta{Title: "集群管理", Icon: "hub"},
		},
	}
	utils.RegisterMenus(entities...)
}
