// Package initialize asset 插件 API 记录种子（幂等，写入 sys_apis）
package initialize

import (
	"context"

	model "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
)

// Api 注册资产中心 API 记录（供角色-权限分配界面展示）
func Api(ctx context.Context) {
	entities := []model.SysApi{
		{Path: "/asset/host/create", Description: "创建主机资产", ApiGroup: "资产管理", Method: "POST"},
		{Path: "/asset/host/delete", Description: "删除主机资产", ApiGroup: "资产管理", Method: "DELETE"},
		{Path: "/asset/host/deleteByIds", Description: "批量删除主机资产", ApiGroup: "资产管理", Method: "DELETE"},
		{Path: "/asset/host/update", Description: "更新主机资产", ApiGroup: "资产管理", Method: "PUT"},
		{Path: "/asset/host/find", Description: "根据ID获取主机资产", ApiGroup: "资产管理", Method: "GET"},
		{Path: "/asset/host/list", Description: "分页获取主机资产列表", ApiGroup: "资产管理", Method: "POST"},
		{Path: "/asset/room/create", Description: "创建机房", ApiGroup: "资产管理", Method: "POST"},
		{Path: "/asset/room/delete", Description: "删除机房", ApiGroup: "资产管理", Method: "DELETE"},
		{Path: "/asset/room/update", Description: "更新机房", ApiGroup: "资产管理", Method: "PUT"},
		{Path: "/asset/room/list", Description: "机房全量列表", ApiGroup: "资产管理", Method: "GET"},
		{Path: "/asset/rack/create", Description: "创建机柜", ApiGroup: "资产管理", Method: "POST"},
		{Path: "/asset/rack/delete", Description: "删除机柜", ApiGroup: "资产管理", Method: "DELETE"},
		{Path: "/asset/rack/update", Description: "更新机柜", ApiGroup: "资产管理", Method: "PUT"},
		{Path: "/asset/rack/list", Description: "机柜列表", ApiGroup: "资产管理", Method: "GET"},
		{Path: "/asset/productLine/create", Description: "创建产品线", ApiGroup: "资产管理", Method: "POST"},
		{Path: "/asset/productLine/delete", Description: "删除产品线", ApiGroup: "资产管理", Method: "DELETE"},
		{Path: "/asset/productLine/update", Description: "更新产品线", ApiGroup: "资产管理", Method: "PUT"},
		{Path: "/asset/productLine/list", Description: "产品线全量列表", ApiGroup: "资产管理", Method: "GET"},
	}
	utils.RegisterApis(entities...)
}
