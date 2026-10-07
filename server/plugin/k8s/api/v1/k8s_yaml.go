// Package api K2 收官：YAML 编辑下发接口（diff 预览 / apply 下发，写级权限仅 888）
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
)

type k8sYaml struct{}

// workloadYamlReq YAML 下发请求体
type workloadYamlReq struct {
	Kind      string `json:"kind" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
	YAML      string `json:"yaml" binding:"required"`
}

// PreviewWorkloadYAML YAML 变更预览（服务端 dry-run diff）
// @Tags K8sYaml
// @Summary 工作负载 YAML 变更预览（server dry-run，返回与当前对象的行 diff）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param body body workloadYamlReq true "工作负载 YAML"
// @Success 200 {object} response.Response{data=string} "diff 文本"
// @Router /k8s/workload/diff [post]
func (y *k8sYaml) PreviewWorkloadYAML(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	var req workloadYamlReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数无效: "+err.Error(), c)
		return
	}
	diff, err := k8sClusterService.PreviewWorkloadYAML(clusterID, req.Kind, req.Namespace, req.Name, req.YAML)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(diff, c)
}

// ApplyWorkloadYAML YAML 下发
// @Tags K8sYaml
// @Summary 工作负载 YAML 下发（乐观锁更新；写操作仅 888）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param body body workloadYamlReq true "工作负载 YAML"
// @Success 200 {object} response.Response{msg=string} "下发成功"
// @Router /k8s/workload/apply [post]
func (y *k8sYaml) ApplyWorkloadYAML(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	var req workloadYamlReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数无效: "+err.Error(), c)
		return
	}
	if err := k8sClusterService.ApplyWorkloadYAML(clusterID, req.Kind, req.Namespace, req.Name, req.YAML); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("YAML 已下发（滚动生效视工作负载策略）", c)
}
