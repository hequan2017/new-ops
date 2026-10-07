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
		{Path: "/k8s/cluster/overview", Description: "集群总览", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/pod/list", Description: "Pod 列表", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/pod/logs", Description: "Pod 日志", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/pod/detail", Description: "Pod 详情", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/pod/delete", Description: "删除 Pod", ApiGroup: "K8s管理", Method: "POST"},
		{Path: "/k8s/pod/execws", Description: "Pod WebShell WebSocket", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/workload/diff", Description: "工作负载 YAML 变更预览", ApiGroup: "K8s管理", Method: "POST"},
		{Path: "/k8s/workload/apply", Description: "工作负载 YAML 下发", ApiGroup: "K8s管理", Method: "POST"},
		{Path: "/k8s/helm/list", Description: "Helm release 列表", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/helm/history", Description: "Helm release 历史", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/helm/install", Description: "Helm 安装", ApiGroup: "K8s管理", Method: "POST"},
		{Path: "/k8s/helm/uninstall", Description: "Helm 卸载", ApiGroup: "K8s管理", Method: "POST"},
		{Path: "/k8s/helm/rollback", Description: "Helm 回滚", ApiGroup: "K8s管理", Method: "POST"},
		{Path: "/k8s/deployment/list", Description: "Deployment 列表", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/deployment/scale", Description: "Deployment 扩缩容", ApiGroup: "K8s管理", Method: "POST"},
		{Path: "/k8s/deployment/restart", Description: "Deployment 滚动重启", ApiGroup: "K8s管理", Method: "POST"},
		{Path: "/k8s/statefulset/list", Description: "StatefulSet 列表", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/daemonset/list", Description: "DaemonSet 列表", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/workload/yaml", Description: "工作负载 YAML 查看", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/node/list", Description: "Node 列表", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/node/detail", Description: "Node 详情", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/node/cordon", Description: "Node 隔离/恢复调度", ApiGroup: "K8s管理", Method: "POST"},
		{Path: "/k8s/node/drain", Description: "Node 驱逐", ApiGroup: "K8s管理", Method: "POST"},
		{Path: "/k8s/service/list", Description: "Service 列表", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/configmap/list", Description: "ConfigMap 列表", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/secret/list", Description: "Secret 列表", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/pvc/list", Description: "PVC 列表", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/ingress/list", Description: "Ingress 列表", ApiGroup: "K8s管理", Method: "GET"},
		{Path: "/k8s/event/list", Description: "事件列表", ApiGroup: "K8s管理", Method: "GET"},
	}
	utils.RegisterApis(entities...)
}
