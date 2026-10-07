// Package initialize k8s 插件 API 记录种子
package initialize

import (
	"context"

	model "github.com/hequan2017/new-ops/server/model/system"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/utils"
)

// Api 注册 k8s API 记录
func Api(ctx context.Context) {
	entities := []model.SysApi{
		{Path: "/k8s/cluster/create", Description: "注册集群", ApiGroup: "K8s管理", Method: "POST"},
		{Path: "/k8s/cluster/delete", Description: "删除集群", ApiGroup: "K8s管理", Method: "DELETE"},
		{Path: "/k8s/cluster/list", Description: "集群列表", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/cluster/test", Description: "集群连接测试", ApiGroup: "K8s管理", Method: "GET"},
	}
	utils.RegisterApis(entities...)
}
