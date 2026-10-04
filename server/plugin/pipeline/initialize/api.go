// Package initialize 白泽流水线 API 记录种子
package initialize

import (
	"context"

	model "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
)

// Api 注册流水线 API 记录
func Api(ctx context.Context) {
	entities := []model.SysApi{
		{Path: "/pipeline", Description: "创建流水线", ApiGroup: "流水线", Method: "POST"},
		{Path: "/pipeline", Description: "更新流水线", ApiGroup: "流水线", Method: "PUT"},
		{Path: "/pipeline", Description: "删除流水线", ApiGroup: "流水线", Method: "DELETE"},
		{Path: "/pipeline/list", Description: "流水线列表", ApiGroup: "流水线", Method: "GET"},
		{Path: "/pipeline/find", Description: "流水线详情", ApiGroup: "流水线", Method: "GET"},
		{Path: "/pipeline/build/start", Description: "触发构建", ApiGroup: "流水线", Method: "POST"},
		{Path: "/pipeline/build/cancel", Description: "取消构建", ApiGroup: "流水线", Method: "POST"},
		{Path: "/pipeline/build/approve", Description: "审批放行", ApiGroup: "流水线", Method: "POST"},
		{Path: "/pipeline/build/list", Description: "构建分页列表", ApiGroup: "流水线", Method: "POST"},
		{Path: "/pipeline/build/logs", Description: "构建日志", ApiGroup: "流水线", Method: "GET"},
	}
	utils.RegisterApis(entities...)
}
