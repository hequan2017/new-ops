// Package initialize 白泽批量作业菜单种子
package initialize

import (
	"context"

	model "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
)

// Menu 注册批量作业菜单
func Menu(ctx context.Context) {
	entities := []model.SysBaseMenu{
		{
			Path:      "job",
			Name:      "job",
			Hidden:    false,
			Component: "view/routerHolder.vue",
			Sort:      4,
			Meta:      model.Meta{Title: "批量作业", Icon: "suitcase"},
		},
		{
			Path:      "jobBatchExec",
			Name:      "jobBatchExec",
			Hidden:    false,
			Component: "plugin/job/view/batchExec/index.vue",
			Sort:      1,
			Meta:      model.Meta{Title: "批量执行", Icon: "video-play"},
		},
		{
			Path:      "jobScript",
			Name:      "jobScript",
			Hidden:    false,
			Component: "plugin/job/view/script/index.vue",
			Sort:      2,
			Meta:      model.Meta{Title: "脚本库", Icon: "document"},
		},
		{
			Path:      "jobVarGroup",
			Name:      "jobVarGroup",
			Hidden:    false,
			Component: "plugin/job/view/varGroup/index.vue",
			Sort:      3,
			Meta:      model.Meta{Title: "变量组", Icon: "key"},
		},
	}
	utils.RegisterMenus(entities...)
}
