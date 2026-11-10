// Package v1 eduadmin API：全场景数据接入（需求 3.6-1）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// eduImportReq 导入请求
type eduImportReq struct {
	Format  string `json:"format" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// SaveEduRecord 保存记录
// @Tags Eduadmin
// @Summary 新建/更新全场景数据记录（唯一键 domain+externalKey）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduRecord true "记录"
// @Success 200 {object} response.Response{data=model.EduRecord} "保存成功"
// @Router /eduadmin/record [post]
func (a *EduadminApi) SaveEduRecord(c *gin.Context) {
	var in model.EduRecord
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduRecord(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// ImportEduRecords 批量导入
// @Tags Eduadmin
// @Summary 批量导入数据记录（JSON/CSV 逐行校验+唯一键去重+错误行报告）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.eduImportReq true "导入请求"
// @Success 200 {object} response.Response{data=map[string]any} "导入完成"
// @Router /eduadmin/record/import [post]
func (a *EduadminApi) ImportEduRecords(c *gin.Context) {
	var in eduImportReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.ImportEduRecords(in.Format, in.Content)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "导入完成", c)
}

// GetEduRecords 记录列表
// @Tags Eduadmin
// @Summary 数据记录列表（域/来源过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param domain query string false "人事/财务/资产/后勤"
// @Param source query string false "来源关键字"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/record/list [get]
func (a *EduadminApi) GetEduRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.EduadminApp.GetEduRecords(c.Query("domain"), c.Query("source"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// DeleteEduRecord 删除记录
// @Tags Eduadmin
// @Summary 删除数据记录（软删除）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "记录ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /eduadmin/record [delete]
func (a *EduadminApi) DeleteEduRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.EduadminApp.DeleteEduRecord(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
