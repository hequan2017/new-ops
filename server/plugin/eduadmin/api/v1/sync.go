// Package v1 eduadmin API：数据同步接口（需求 3.6-29~33）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// saveAdapterReq 适配器请求（凭据明文仅入口，密文落库不回显）
type saveAdapterReq struct {
	model.EduSyncAdapter
	Credential string `json:"credential"`
}

// SaveEduSyncAdapter 保存适配器
// @Tags Eduadmin
// @Summary 新建/更新同步适配器（凭据 AES-GCM 密文落库，同步频率 ≤5 分钟口径）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.saveAdapterReq true "适配器"
// @Success 200 {object} response.Response{data=model.EduSyncAdapter} "保存成功"
// @Router /eduadmin/sync/adapter [post]
func (a *EduadminApi) SaveEduSyncAdapter(c *gin.Context) {
	var in saveAdapterReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduSyncAdapter(&in.EduSyncAdapter, in.Credential)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetEduSyncAdapters 适配器列表
// @Tags Eduadmin
// @Summary 同步适配器列表（凭据不回显）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.EduSyncAdapter} "获取成功"
// @Router /eduadmin/sync/adapter/list [get]
func (a *EduadminApi) GetEduSyncAdapters(c *gin.Context) {
	list, err := service.EduadminApp.GetEduSyncAdapters()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// RunEduSyncTask 执行同步
// @Tags Eduadmin
// @Summary 执行同步任务（HTTP 拉取→域/唯一键校验→幂等落库→执行记录）
// @Security ApiKeyAuth
// @Produce application/json
// @Param adapterId query int true "适配器ID"
// @Success 200 {object} response.Response{data=model.EduSyncTask} "执行完成"
// @Router /eduadmin/sync/run [post]
func (a *EduadminApi) RunEduSyncTask(c *gin.Context) {
	adapterId, _ := strconv.ParseUint(c.Query("adapterId"), 10, 64)
	out, err := service.EduadminApp.RunEduSyncTask(uint(adapterId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "执行完成", c)
}

// GetEduSyncTasks 任务记录
// @Tags Eduadmin
// @Summary 同步任务执行记录（适配器过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param adapterId query int false "适配器ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/sync/task/list [get]
func (a *EduadminApi) GetEduSyncTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	adapterId, _ := strconv.ParseUint(c.Query("adapterId"), 10, 64)
	list, total, err := service.EduadminApp.GetEduSyncTasks(uint(adapterId), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}
