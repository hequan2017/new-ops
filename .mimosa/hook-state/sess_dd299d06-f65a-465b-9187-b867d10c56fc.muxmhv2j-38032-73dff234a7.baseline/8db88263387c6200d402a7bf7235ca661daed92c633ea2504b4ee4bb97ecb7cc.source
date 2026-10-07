// Package api K2 资源浏览接口（只读，888 全部 / 9528 只读）
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
)

type k8sResource struct{}

// parseUintQ uint query 解析
func parseUintQ(c *gin.Context, name string) (uint, bool) {
	v, err := strconv.ParseUint(c.Query(name), 10, 64)
	if err != nil || v == 0 {
		response.FailWithMessage(name+" 无效", c)
		return 0, false
	}
	return uint(v), true
}

// ListPods Pod 列表
// @Tags K8sResource
// @Summary Pod 列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param namespace query string false "命名空间（空=全部）"
// @Success 200 {object} response.Response{data=[]service.PodInfo} "获取成功"
// @Router /k8s/pod/list [get]
func (t *k8sResource) ListPods(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	list, err := k8sClusterService.ListPods(clusterID, c.Query("namespace"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// GetPodLogs Pod 日志
// @Tags K8sResource
// @Summary Pod 日志
// @Security ApiKeyAuth
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Param pod query string true "Pod 名"
// @Param container query string false "容器名（多容器时必填）"
// @Param tailLines query int false "尾部行数（默认500，上限2000）"
// @Success 200 {object} response.Response{data=string} "获取成功"
// @Router /k8s/pod/logs [get]
func (t *k8sResource) GetPodLogs(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	ns := c.Query("namespace")
	pod := c.Query("pod")
	if ns == "" || pod == "" {
		response.FailWithMessage("namespace/pod 必填", c)
		return
	}
	tail, _ := strconv.ParseInt(c.DefaultQuery("tailLines", "500"), 10, 64)
	logs, err := k8sClusterService.GetPodLogs(clusterID, ns, pod, c.Query("container"), tail)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(logs, c)
}

// ListDeployments Deployment 列表
// @Tags K8sResource
// @Summary Deployment 列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param namespace query string false "命名空间（空=全部）"
// @Success 200 {object} response.Response{data=[]service.DeploymentInfo} "获取成功"
// @Router /k8s/deployment/list [get]
func (t *k8sResource) ListDeployments(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	list, err := k8sClusterService.ListDeployments(clusterID, c.Query("namespace"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// ListNodes 节点列表
// @Tags K8sResource
// @Summary 节点列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Success 200 {object} response.Response{data=[]service.NodeInfo} "获取成功"
// @Router /k8s/node/list [get]
func (t *k8sResource) ListNodes(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	list, err := k8sClusterService.ListNodes(clusterID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}
