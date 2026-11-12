// Package v1 artsci API：文理学科 AI 工具集接口层（参数绑定与响应，禁止业务逻辑）
package v1

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"

	"github.com/hequan2017/new-ops/server/plugin/artsci/model"
	"github.com/hequan2017/new-ops/server/plugin/artsci/service"
)

var Api = new(ArtsciApi)

type ArtsciApi struct{}

// runToolReq 执行工具请求
type runToolReq struct {
	ID       uint   `json:"id" binding:"required"`
	ArgsJSON string `json:"argsJSON"`
}

// Ping 插件连通性探测
// @Tags Artsci
// @Summary artsci 连通性探测
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]any} "探测成功"
// @Router /artsci/ping [get]
func (a *ArtsciApi) Ping(c *gin.Context) {
	response.OkWithDetailed(gin.H{
		"plugin": "artsci",
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}, "探测成功", c)
}

// GetTools 工具列表
// @Tags Artsci
// @Summary 学科工具列表（学科/关键字过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param subject query string false "学科(数学/文科/图像)"
// @Param keyword query string false "名称/描述关键字"
// @Success 200 {object} response.Response{data=[]model.ArtsciTool} "获取成功"
// @Router /artsci/tool/list [get]
func (a *ArtsciApi) GetTools(c *gin.Context) {
	list, err := service.ArtsciApp.GetTools(c.Query("subject"), c.Query("keyword"))
	if err != nil {
		errx.Respond(c, err, "获取工具列表失败")
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveTool 保存工具
// @Tags Artsci
// @Summary 新建/更新学科工具
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.ArtsciTool true "工具"
// @Success 200 {object} response.Response{data=model.ArtsciTool} "保存成功"
// @Router /artsci/tool [post]
func (a *ArtsciApi) SaveTool(c *gin.Context) {
	var in model.ArtsciTool
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.ArtsciApp.SaveTool(&in)
	if err != nil {
		errx.Respond(c, err, "保存工具失败")
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeleteTool 删除工具
// @Tags Artsci
// @Summary 删除学科工具（软删除）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "工具ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /artsci/tool [delete]
func (a *ArtsciApi) DeleteTool(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.ArtsciApp.DeleteTool(uint(id)); err != nil {
		errx.Respond(c, err, "删除工具失败")
		return
	}
	response.OkWithMessage("删除成功", c)
}

// RunTool 执行工具
// @Tags Artsci
// @Summary 执行学科工具（module 分发内置实现，成败均留痕）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.runToolReq true "执行请求"
// @Success 200 {object} response.Response{data=model.ArtsciRunLog} "执行成功"
// @Router /artsci/tool/run [post]
func (a *ArtsciApi) RunTool(c *gin.Context) {
	var in runToolReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	log, err := service.ArtsciApp.RunTool(in.ID, in.ArgsJSON)
	if err != nil {
		if log != nil {
			response.OkWithDetailed(log, "执行失败（已留痕）: "+err.Error(), c)
			return
		}
		errx.Respond(c, err, "执行工具失败")
		return
	}
	response.OkWithDetailed(log, "执行成功", c)
}

// GetRunLogs 执行留痕
// @Tags Artsci
// @Summary 工具执行留痕列表（工具过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param toolId query int false "工具ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /artsci/tool/runLogs [get]
func (a *ArtsciApi) GetRunLogs(c *gin.Context) {
	toolId, _ := strconv.ParseUint(c.Query("toolId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ArtsciApp.GetRunLogs(uint(toolId), page, pageSize)
	if err != nil {
		errx.Respond(c, err, "查询执行留痕失败")
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GradeEssay 智能作业批改
// @Tags Artsci
// @Summary 提交作业批改（llm 结构化评分/规则口径 fallback，成败均落记录）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.ArtsciGrading true "批改请求"
// @Success 200 {object} response.Response{data=model.ArtsciGrading} "批改完成"
// @Router /artsci/grading [post]
func (a *ArtsciApi) GradeEssay(c *gin.Context) {
	var in model.ArtsciGrading
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	rec, err := service.ArtsciApp.GradeEssay(&in)
	if err != nil {
		response.OkWithDetailed(rec, "批改失败（已留痕）: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(rec, "批改完成", c)
}

// GetGradings 批改记录列表
// @Tags Artsci
// @Summary 作业批改记录列表（倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /artsci/grading/list [get]
func (a *ArtsciApi) GetGradings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ArtsciApp.GetGradings(page, pageSize)
	if err != nil {
		errx.Respond(c, err, "查询批改记录失败")
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// genQuestionReq 出题请求
type genQuestionReq struct {
	KnowledgePoint string  `json:"knowledgePoint" binding:"required"`
	QType          string  `json:"qType" binding:"required"`
	Difficulty     string  `json:"difficulty"`
	ModelId        uint    `json:"modelId"`
	Score          float64 `json:"score"`
}

// buildPaperReq 组卷请求
type buildPaperReq struct {
	Name      string `json:"name" binding:"required"`
	RulesJSON string `json:"rulesJSON" binding:"required"`
}

// GenerateQuestion 自动出题
// @Tags Artsci
// @Summary 生成题目（模板规则+chat 可选增强，失败回落模板）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.genQuestionReq true "出题请求"
// @Success 200 {object} response.Response{data=model.ArtsciQuestion} "生成成功"
// @Router /artsci/question/generate [post]
func (a *ArtsciApi) GenerateQuestion(c *gin.Context) {
	var in genQuestionReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	q, err := service.ArtsciApp.GenerateQuestion(in.KnowledgePoint, in.QType, in.Difficulty, in.ModelId, in.Score)
	if err != nil {
		errx.Respond(c, err, "生成题目失败")
		return
	}
	response.OkWithDetailed(q, "生成成功", c)
}

// GetQuestions 题目列表
// @Tags Artsci
// @Summary 生成题目列表（知识点/题型过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param knowledgePoint query string false "知识点"
// @Param qType query string false "题型"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /artsci/question/list [get]
func (a *ArtsciApi) GetQuestions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ArtsciApp.GetQuestions(c.Query("knowledgePoint"), c.Query("qType"), page, pageSize)
	if err != nil {
		errx.Respond(c, err, "查询题目列表失败")
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// BuildPaper 规则组卷
// @Tags Artsci
// @Summary 按规则抽题组卷（知识点×题型×题量×单题分，分值合计服务端计算）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.buildPaperReq true "组卷请求"
// @Success 200 {object} response.Response{data=model.ArtsciPaper} "组卷成功"
// @Router /artsci/paper [post]
func (a *ArtsciApi) BuildPaper(c *gin.Context) {
	var in buildPaperReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	paper, err := service.ArtsciApp.BuildPaper(in.Name, in.RulesJSON)
	if err != nil {
		errx.Respond(c, err, "组卷失败")
		return
	}
	response.OkWithDetailed(paper, "组卷成功", c)
}

// GetPapers 试卷列表
// @Tags Artsci
// @Summary 试卷列表（倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /artsci/paper/list [get]
func (a *ArtsciApi) GetPapers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ArtsciApp.GetPapers(page, pageSize)
	if err != nil {
		errx.Respond(c, err, "查询试卷列表失败")
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetSceneTemplates 场景模板列表
// @Tags Artsci
// @Summary 文生图教学场景模板列表（五类 seed）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.ArtsciSceneTemplate} "获取成功"
// @Router /artsci/scene/list [get]
func (a *ArtsciApi) GetSceneTemplates(c *gin.Context) {
	list, err := service.ArtsciApp.GetSceneTemplates()
	if err != nil {
		errx.Respond(c, err, "查询场景模板失败")
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// submitImageTaskReq 文生图任务请求
type submitImageTaskReq struct {
	Scene     string `json:"scene" binding:"required"`
	InputText string `json:"inputText" binding:"required"`
}

// SubmitImageTask 提交文生图任务
// @Tags Artsci
// @Summary 提交文生图任务（模板渲染→适配器生成，状态机 queued/done/failed）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.submitImageTaskReq true "任务请求"
// @Success 200 {object} response.Response{data=model.ArtsciImageTask} "任务完成"
// @Router /artsci/imageTask [post]
func (a *ArtsciApi) SubmitImageTask(c *gin.Context) {
	var in submitImageTaskReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	task, err := service.ArtsciApp.SubmitImageTask(in.Scene, in.InputText)
	if err != nil {
		errx.Respond(c, err, "提交文生图任务失败")
		return
	}
	response.OkWithDetailed(task, "任务完成", c)
}

// GetImageTasks 文生图任务列表
// @Tags Artsci
// @Summary 文生图任务列表（场景过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param scene query string false "场景"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /artsci/imageTask/list [get]
func (a *ArtsciApi) GetImageTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ArtsciApp.GetImageTasks(c.Query("scene"), page, pageSize)
	if err != nil {
		errx.Respond(c, err, "查询文生图任务列表失败")
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// RegisterImageAsset 图像资产登记
// @Tags Artsci
// @Summary 登记图像资产（路径/MIME/来源元数据）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.ArtsciImageAsset true "图像资产"
// @Success 200 {object} response.Response{data=model.ArtsciImageAsset} "登记成功"
// @Router /artsci/imageAsset [post]
func (a *ArtsciApi) RegisterImageAsset(c *gin.Context) {
	var in model.ArtsciImageAsset
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.ArtsciApp.RegisterImageAsset(&in)
	if err != nil {
		errx.Respond(c, err, "登记图像资产失败")
		return
	}
	response.OkWithDetailed(out, "登记成功", c)
}

// GetImageAssets 图像资产列表
// @Tags Artsci
// @Summary 图像资产列表（关键字过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param keyword query string false "名称关键字"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /artsci/imageAsset/list [get]
func (a *ArtsciApi) GetImageAssets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ArtsciApp.GetImageAssets(c.Query("keyword"), page, pageSize)
	if err != nil {
		errx.Respond(c, err, "查询图像资产列表失败")
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// interpretReq 图像解读请求
type interpretReq struct {
	AssetId uint   `json:"assetId" binding:"required"`
	Subject string `json:"subject" binding:"required"`
	ModelId uint   `json:"modelId"`
}

// InterpretImage 图像解读
// @Tags Artsci
// @Summary 图像解读与分析（表面细节/文化含义两段结构化，llm 或元数据口径）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.interpretReq true "解读请求"
// @Success 200 {object} response.Response{data=model.ArtsciInterpretation} "解读完成"
// @Router /artsci/imageAsset/interpret [post]
func (a *ArtsciApi) InterpretImage(c *gin.Context) {
	var in interpretReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	rec, err := service.ArtsciApp.InterpretImage(in.AssetId, in.Subject, in.ModelId)
	if err != nil {
		if rec != nil {
			response.OkWithDetailed(rec, "解读失败（已留痕）: "+err.Error(), c)
			return
		}
		errx.Respond(c, err, "图像解读失败")
		return
	}
	response.OkWithDetailed(rec, "解读完成", c)
}

// GetInterpretations 解读记录列表
// @Tags Artsci
// @Summary 图像解读记录列表（资产过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param assetId query int false "资产ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /artsci/interpretation/list [get]
func (a *ArtsciApi) GetInterpretations(c *gin.Context) {
	assetId, _ := strconv.ParseUint(c.Query("assetId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ArtsciApp.GetInterpretations(uint(assetId), page, pageSize)
	if err != nil {
		errx.Respond(c, err, "查询解读记录失败")
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetToolRunStats 执行统计
// @Tags Artsci
// @Summary 工具执行留痕统计（按工具聚合 成功率/均耗时，近 2000 条）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]service.toolStat} "获取成功"
// @Router /artsci/runLog/stats [get]
func (a *ArtsciApi) GetToolRunStats(c *gin.Context) {
	out, err := service.ArtsciApp.GetToolRunStats()
	if err != nil {
		errx.Respond(c, err, "查询执行统计失败")
		return
	}
	response.OkWithDetailed(out, "获取成功", c)
}
