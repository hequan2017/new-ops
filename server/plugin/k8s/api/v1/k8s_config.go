// Package api K2 补充：Service/ConfigMap/Secret 只读接口
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hequan2017/new-ops/server/model/common/response"
)

type k8sConfig struct{}

// ListServices Service 列表
// @Tags K8sResource
// @Summary Service 列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param namespace query string false "命名空间（空=全部）"
// @Success 200 {object} response.Response{data=[]service.ServiceInfo} "获取成功"
// @Router /k8s/service/list [get]
func (t *k8sResource) ListServices(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	list, err := k8sClusterService.ListServicesScoped(clusterID, c.Query("namespace"), scopeUID(c), scopeAuth(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// ListConfigMaps ConfigMap 列表（只回显数据键）
// @Tags K8sResource
// @Summary ConfigMap 列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param namespace query string false "命名空间（空=全部）"
// @Success 200 {object} response.Response{data=[]service.ConfigMapInfo} "获取成功"
// @Router /k8s/configmap/list [get]
func (t *k8sResource) ListConfigMaps(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	list, err := k8sClusterService.ListConfigMapsScoped(clusterID, c.Query("namespace"), scopeUID(c), scopeAuth(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// ListSecrets Secret 列表（值永不回显）
// @Tags K8sResource
// @Summary Secret 列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param clusterId query int true "集群ID"
// @Param namespace query string false "命名空间（空=全部）"
// @Success 200 {object} response.Response{data=[]service.SecretInfo} "获取成功"
// @Router /k8s/secret/list [get]
func (t *k8sResource) ListSecrets(c *gin.Context) {
	clusterID, ok := parseUintQ(c, "clusterId")
	if !ok {
		return
	}
	list, err := k8sClusterService.ListSecretsScoped(clusterID, c.Query("namespace"), scopeUID(c), scopeAuth(c))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}
