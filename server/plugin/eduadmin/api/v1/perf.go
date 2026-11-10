// Package v1 eduadmin API：绩效考评（需求 3.6-13）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SaveEduPerfRecord 保存绩效数据
// @Tags Eduadmin
// @Summary 新建/更新绩效数据（教师/行政权重口径）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduPerfRecord true "绩效数据"
// @Success 200 {object} response.Response{data=model.EduPerfRecord} "保存成功"
// @Router /eduadmin/perf [post]
func (a *EduadminApi) SaveEduPerfRecord(c *gin.Context) {
	var in model.EduPerfRecord
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduPerfRecord(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetEduPerfRecords 绩效数据列表
// @Tags Eduadmin
// @Summary 绩效数据列表（姓名/年度过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param staffName query string false "姓名"
// @Param year query int false "年度"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/perf/list [get]
func (a *EduadminApi) GetEduPerfRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	year, _ := strconv.Atoi(c.Query("year"))
	list, total, err := service.EduadminApp.GetEduPerfRecords(c.Query("staffName"), year, page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// perfReportReq 报告生成请求
type perfReportReq struct {
	StaffName string `json:"staffName" binding:"required"`
	Year      int    `json:"year" binding:"required"`
}

// GeneratePerfReport 生成报告
// @Tags Eduadmin
// @Summary 生成个性化绩效报告（模板+数据填充，同年同人幂等覆盖）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.perfReportReq true "报告请求"
// @Success 200 {object} response.Response{data=model.EduPerfReport} "生成成功"
// @Router /eduadmin/perf/report [post]
func (a *EduadminApi) GeneratePerfReport(c *gin.Context) {
	var in perfReportReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.GeneratePerfReport(in.StaffName, in.Year)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "生成成功", c)
}

// GetEduPerfReports 报告列表
// @Tags Eduadmin
// @Summary 绩效报告列表（姓名/年度过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param staffName query string false "姓名"
// @Param year query int false "年度"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/perf/report/list [get]
func (a *EduadminApi) GetEduPerfReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	year, _ := strconv.Atoi(c.Query("year"))
	list, total, err := service.EduadminApp.GetEduPerfReports(c.Query("staffName"), year, page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}
