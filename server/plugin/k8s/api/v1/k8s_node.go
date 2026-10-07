// Package api K1 Node 管理与集群总览接口
package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
)

type k8sNode struct{}

// parseClusterIDQuery 集群 ID 解析（query clusterId）
func parseClusterIDQuery(c *gin.Context) (uint, bool) {
	cid, err := strconv.ParseUint(c.Query("clusterId"), 10, 64)
	if err != nil || cid == 0 {
		response.FailWithMessage("clusterId 无效", c)
		return 0, false
	}
	return uint(cid), true
}

// GetClusterOverview 集群总览
// @Tags K8sNode
// @Summary 集群总览（版本/节点/命名空间/Pod/metrics 资源用量）
// @Security ApiKeyAuth
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Success 200 {object} response.Response{data=service.ClusterOverview,msg=string} "总览数据"
// @Router /k8s/cluster/overview [get]
func (n *k8sNode) GetClusterOverview(c *gin.Context) {
	clusterID, ok := parseClusterIDQuery(c)
	if !ok {
		return
	}
	ov, err := k8sClusterService.GetClusterOverview(clusterID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(ov, c)
}

// GetNodeDetail 节点详情
// @Tags K8sNode
// @Summary 节点详情（allocatable/capacity/conditions/taints）
// @Security ApiKeyAuth
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param name query string true "节点名"
// @Success 200 {object} response.Response{data=service.NodeDetail,msg=string} "节点详情"
// @Router /k8s/node/detail [get]
func (n *k8sNode) GetNodeDetail(c *gin.Context) {
	clusterID, ok := parseClusterIDQuery(c)
	if !ok {
		return
	}
	name := c.Query("name")
	if name == "" {
		response.FailWithMessage("name 必填", c)
		return
	}
	d, err := k8sClusterService.GetNodeDetail(clusterID, name)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(d, c)
}

// CordonNode 节点隔离/恢复调度
// @Tags K8sNode
// @Summary 节点 cordon/uncordon（写操作，仅 888）
// @Security ApiKeyAuth
// @Accept application/x-www-form-urlencoded
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param name formData string true "节点名"
// @Param cordon formData bool true "true=隔离 false=恢复调度"
// @Success 200 {object} response.Response{msg=string} "操作成功"
// @Router /k8s/node/cordon [post]
func (n *k8sNode) CordonNode(c *gin.Context) {
	clusterID, ok := parseClusterIDQuery(c)
	if !ok {
		return
	}
	name := c.PostForm("name")
	if name == "" {
		response.FailWithMessage("name 必填", c)
		return
	}
	cordon := c.PostForm("cordon") == "true"
	if err := k8sClusterService.SetNodeSchedulability(clusterID, name, cordon); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if cordon {
		response.OkWithMessage("节点已隔离（禁止新 Pod 调度）", c)
	} else {
		response.OkWithMessage("节点已恢复调度", c)
	}
}

// DrainNode 节点驱逐
// @Tags K8sNode
// @Summary 节点 drain（先隔离再驱逐非 DaemonSet/非镜像 Pod，写操作仅 888）
// @Security ApiKeyAuth
// @Accept application/x-www-form-urlencoded
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param name formData string true "节点名"
// @Param gracePeriod formData int false "宽限期秒数（-1 或缺省用 Pod 默认）"
// @Success 200 {object} response.Response{data=service.DrainResult,msg=string} "驱逐结果"
// @Router /k8s/node/drain [post]
func (n *k8sNode) DrainNode(c *gin.Context) {
	clusterID, ok := parseClusterIDQuery(c)
	if !ok {
		return
	}
	name := c.PostForm("name")
	if name == "" {
		response.FailWithMessage("name 必填", c)
		return
	}
	gracePeriod := int64(-1)
	if g := c.PostForm("gracePeriod"); g != "" {
		if v, err := strconv.ParseInt(g, 10, 64); err == nil {
			gracePeriod = v
		}
	}
	res, err := k8sClusterService.DrainNode(clusterID, name, gracePeriod)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(res, "驱逐已提交", c)
}
