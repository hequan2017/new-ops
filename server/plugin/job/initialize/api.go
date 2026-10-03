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
		{Path: "/job/script", Description: "创建脚本", ApiGroup: "批量作业", Method: "POST"},
		{Path: "/job/script", Description: "更新脚本", ApiGroup: "批量作业", Method: "PUT"},
		{Path: "/job/script", Description: "删除脚本", ApiGroup: "批量作业", Method: "DELETE"},
		{Path: "/job/script/list", Description: "脚本列表", ApiGroup: "批量作业", Method: "GET"},
		{Path: "/job/script/versions", Description: "脚本历史版本", ApiGroup: "批量作业", Method: "GET"},
		{Path: "/job/vargroup", Description: "创建变量组", ApiGroup: "批量作业", Method: "POST"},
		{Path: "/job/vargroup", Description: "更新变量组", ApiGroup: "批量作业", Method: "PUT"},
		{Path: "/job/vargroup", Description: "删除变量组", ApiGroup: "批量作业", Method: "DELETE"},
		{Path: "/job/vargroup/list", Description: "变量组列表", ApiGroup: "批量作业", Method: "GET"},
	}
	utils.RegisterApis(entities...)
}
