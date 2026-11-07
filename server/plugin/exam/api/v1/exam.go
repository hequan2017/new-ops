// Package v1 exam API：题库接口层
package v1

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/exam/model"
	"github.com/hequan2017/new-ops/server/plugin/exam/service"
)

var Api = new(ExamApi)

type ExamApi struct{}

// GetQuestions 题目列表
// @Tags Exam
// @Summary 考题列表（题型/知识点过滤，分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param type query string false "single/multiple/judge/short/practice"
// @Param knowledgePoint query string false "知识点"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /exam/question/list [get]
func (a *ExamApi) GetQuestions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ExamApp.GetQuestions(page, pageSize, c.Query("type"), c.Query("knowledgePoint"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// SaveQuestion 保存题目
// @Tags Exam
// @Summary 新建/更新考题
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.ExamQuestion true "考题"
// @Success 200 {object} response.Response{data=model.ExamQuestion} "保存成功"
// @Router /exam/question [post]
func (a *ExamApi) SaveQuestion(c *gin.Context) {
	var in model.ExamQuestion
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	q, err := service.ExamApp.SaveQuestion(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(q, "保存成功", c)
}

// DeleteQuestion 删除题目
// @Tags Exam
// @Summary 删除考题（软删除）
// @Security ApiKeyAuth
// @Param id query int true "题目ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /exam/question [delete]
func (a *ExamApi) DeleteQuestion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.ExamApp.DeleteQuestion(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ImportQuestions JSONL 导入
// @Tags Exam
// @Summary 考题 JSONL 批量导入（10MB 上限）
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Param file formData file true "JSONL 文件"
// @Success 200 {object} response.Response{data=map[string]any} "导入完成"
// @Router /exam/question/import [post]
func (a *ExamApi) ImportQuestions(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("缺少 file 文件字段", c)
		return
	}
	if fh.Size > 10<<20 {
		response.FailWithMessage("文件超过 10MB 上限", c)
		return
	}
	f, err := fh.Open()
	if err != nil {
		response.FailWithMessage("读取上传文件失败", c)
		return
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		response.FailWithMessage("读取上传文件失败", c)
		return
	}
	imported, skipped, err := service.ExamApp.ImportQuestions(string(content))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"imported": imported, "skipped": skipped}, "导入完成", c)
}

// GetPapers 试卷列表
// @Tags Exam
// @Summary 试卷列表
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]model.ExamPaper} "获取成功"
// @Router /exam/paper/list [get]
func (a *ExamApi) GetPapers(c *gin.Context) {
	list, err := service.ExamApp.GetPapers()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SavePaper 保存试卷
// @Tags Exam
// @Summary 新建/更新试卷（规则即检）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.ExamPaper true "试卷"
// @Success 200 {object} response.Response{data=model.ExamPaper} "保存成功"
// @Router /exam/paper [post]
func (a *ExamApi) SavePaper(c *gin.Context) {
	var in model.ExamPaper
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	p, err := service.ExamApp.SavePaper(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(p, "保存成功", c)
}

// ComposePaper 组卷
// @Tags Exam
// @Summary 按规则抽题组卷（锁定题目与总分）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body composeIn true "组卷请求"
// @Success 200 {object} response.Response{data=model.ExamPaper} "组卷完成"
// @Router /exam/paper/compose [post]
func (a *ExamApi) ComposePaper(c *gin.Context) {
	var in composeIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	p, err := service.ExamApp.ComposePaper(in.ID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(p, "组卷完成", c)
}

// DeletePaper 删除试卷
// @Tags Exam
// @Summary 删除试卷（软删除）
// @Security ApiKeyAuth
// @Param id query int true "试卷ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /exam/paper [delete]
func (a *ExamApi) DeletePaper(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.ExamApp.DeletePaper(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// composeIn 组卷请求
type composeIn struct {
	ID uint `json:"id" binding:"required"`
}
