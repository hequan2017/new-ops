// Package initialize 白泽批量作业 API 记录种子
package initialize

import (
	"context"

	model "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
)

// Api 注册批量作业 API 记录
func Api(ctx context.Context) {
	entities := []model.SysApi{
		{Path: "/job/exec", Description: "发起批量命令执行", ApiGroup: "批量作业", Method: "POST"},
		{Path: "/job/exec/cancel", Description: "取消进行中批次", ApiGroup: "批量作业", Method: "POST"},
		{Path: "/job/exec/list", Description: "批次分页列表", ApiGroup: "批量作业", Method: "POST"},
		{Path: "/job/exec/detail", Description: "批次详情与每主机结果", ApiGroup: "批量作业", Method: "GET"},
	}
	utils.RegisterApis(entities...)
}
