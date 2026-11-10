// Package v1 eduadmin API：智能招聘（需求 3.6-12）
package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
	"github.com/hequan2017/new-ops/server/plugin/eduadmin/service"
)

// SaveEduResume 保存简历
// @Tags Eduadmin
// @Summary 新建/更新简历
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduResume true "简历"
// @Success 200 {object} response.Response{data=model.EduResume} "保存成功"
// @Router /eduadmin/resume [post]
func (a *EduadminApi) SaveEduResume(c *gin.Context) {
	var in model.EduResume
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduResume(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetEduResumes 简历列表
// @Tags Eduadmin
// @Summary 简历列表（倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/resume/list [get]
func (a *EduadminApi) GetEduResumes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.EduadminApp.GetEduResumes(page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// SaveEduPosition 保存岗位
// @Tags Eduadmin
// @Summary 新建/更新招聘岗位
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.EduPosition true "岗位"
// @Success 200 {object} response.Response{data=model.EduPosition} "保存成功"
// @Router /eduadmin/position [post]
func (a *EduadminApi) SaveEduPosition(c *gin.Context) {
	var in model.EduPosition
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.EduadminApp.SaveEduPosition(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetEduPositions 岗位列表
// @Tags Eduadmin
// @Summary 招聘岗位列表（倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /eduadmin/position/list [get]
func (a *EduadminApi) GetEduPositions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.EduadminApp.GetEduPositions(page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// ScreenCandidates 候选人筛选
// @Tags Eduadmin
// @Summary 候选人筛选（关键词重合×经验年限融合分降序，高/中/低三档）
// @Security ApiKeyAuth
// @Produce application/json
// @Param positionId query int true "岗位ID"
// @Success 200 {object} response.Response{data=[]service.recruitMatch} "筛选完成"
// @Router /eduadmin/screen [get]
func (a *EduadminApi) ScreenCandidates(c *gin.Context) {
	positionId, _ := strconv.ParseUint(c.Query("positionId"), 10, 64)
	out, err := service.EduadminApp.ScreenCandidates(uint(positionId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "筛选完成", c)
}
