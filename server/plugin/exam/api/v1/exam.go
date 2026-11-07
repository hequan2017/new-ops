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

// GetExams 考试列表
// @Tags Exam
// @Summary 考试场次列表（状态过滤）
// @Security ApiKeyAuth
// @Param status query string false "draft/published/ongoing/finished"
// @Success 200 {object} response.Response{data=[]model.ExamExam} "获取成功"
// @Router /exam/exam/list [get]
func (a *ExamApi) GetExams(c *gin.Context) {
	list, err := service.ExamApp.GetExams(c.Query("status"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveExam 保存考试
// @Tags Exam
// @Summary 新建/更新考试场次（试卷须已组卷，仅草稿可编辑）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.ExamIn true "考试"
// @Success 200 {object} response.Response{data=model.ExamExam} "保存成功"
// @Router /exam/exam [post]
func (a *ExamApi) SaveExam(c *gin.Context) {
	var in service.ExamIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	e, err := service.ExamApp.SaveExam(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(e, "保存成功", c)
}

// TransitionExam 考试状态流转
// @Tags Exam
// @Summary 考试状态流转（四态状态机）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body examTransitionIn true "流转请求"
// @Success 200 {object} response.Response{data=model.ExamExam} "流转成功"
// @Router /exam/exam/transition [post]
func (a *ExamApi) TransitionExam(c *gin.Context) {
	var in examTransitionIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	e, err := service.ExamApp.TransitionExam(in.ID, in.To)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(e, "流转成功", c)
}

// DeleteExam 删除考试
// @Tags Exam
// @Summary 删除考试（发布/进行中拒绝）
// @Security ApiKeyAuth
// @Param id query int true "考试ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /exam/exam [delete]
func (a *ExamApi) DeleteExam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.ExamApp.DeleteExam(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// examTransitionIn 考试流转请求
type examTransitionIn struct {
	ID uint   `json:"id" binding:"required"`
	To string `json:"to" binding:"required"`
}

// GetCandidates 考生列表
// @Tags Exam
// @Summary 考生列表（考试过滤，分页）
// @Security ApiKeyAuth
// @Param examId query int false "考试ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /exam/candidate/list [get]
func (a *ExamApi) GetCandidates(c *gin.Context) {
	examId, _ := strconv.ParseUint(c.Query("examId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ExamApp.GetCandidates(uint(examId), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// ImportCandidates 考生名单导入
// @Tags Exam
// @Summary 名单导入（自动发券，重名跳过）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.CandImportIn true "导入请求"
// @Success 200 {object} response.Response{data=map[string]any} "导入完成"
// @Router /exam/candidate/import [post]
func (a *ExamApi) ImportCandidates(c *gin.Context) {
	var in service.CandImportIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	imported, skipped, err := service.ExamApp.ImportCandidates(in.ExamId, in.Names)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"imported": imported, "skipped": skipped}, "导入完成", c)
}

// ClaimTicket 发券
// @Tags Exam
// @Summary 考生自助领券（行锁幂等，重复领取返回已有券码）
// @Accept application/json
// @Param data body service.ClaimIn true "领取请求"
// @Success 200 {object} response.Response{data=model.ExamCandidate} "领取成功"
// @Router /exam/candidate/claim [post]
func (a *ExamApi) ClaimTicket(c *gin.Context) {
	var in service.ClaimIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	cand, err := service.ExamApp.ClaimTicket(in.ExamId, in.UserKey, in.Name)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(cand, "领取成功", c)
}

// DeleteCandidate 删除考生
// @Tags Exam
// @Summary 删除考生
// @Security ApiKeyAuth
// @Param id query int true "考生ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /exam/candidate [delete]
func (a *ExamApi) DeleteCandidate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.ExamApp.DeleteCandidate(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// LoginByTicket 券码登录
// @Tags Exam
// @Summary 考生券码登录（幂等创建答卷）
// @Accept application/json
// @Param data body service.TicketLoginIn true "登录请求"
// @Success 200 {object} response.Response{data=map[string]any} "登录成功"
// @Router /exam/answer/login [post]
func (a *ExamApi) LoginByTicket(c *gin.Context) {
	var in service.TicketLoginIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	sheet, exam, err := service.ExamApp.LoginByTicket(in.Ticket, in.UserKey)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"sheet": sheet, "exam": gin.H{"id": exam.ID, "name": exam.Name, "durationMin": exam.DurationMin, "endTime": exam.EndTime}}, "登录成功", c)
}

// GetSheetQuestions 拉题
// @Tags Exam
// @Summary 拉取答卷题目（标准答案脱敏）与已存答案
// @Security ApiKeyAuth
// @Param sheetId query int true "答卷ID"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /exam/answer/questions [get]
func (a *ExamApi) GetSheetQuestions(c *gin.Context) {
	sheetId, err := strconv.ParseUint(c.Query("sheetId"), 10, 64)
	if err != nil || sheetId == 0 {
		response.FailWithMessage("sheetId 不能为空", c)
		return
	}
	qs, sheet, answers, err := service.ExamApp.GetSheetQuestions(uint(sheetId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"questions": qs, "sheet": sheet, "answers": answers}, "获取成功", c)
}

// SaveAnswers 保存答案
// @Tags Exam
// @Summary 幂等保存答案（乐观锁版本 + 时间窗校验）
// @Accept application/json
// @Param data body service.SaveAnswersIn true "保存请求"
// @Success 200 {object} response.Response{data=model.ExamAnswerSheet} "保存成功"
// @Router /exam/answer/save [post]
func (a *ExamApi) SaveAnswers(c *gin.Context) {
	var in service.SaveAnswersIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	sheet, err := service.ExamApp.SaveAnswers(in.SheetId, in.Answers, in.Version)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(sheet, "保存成功", c)
}

// SubmitSheet 交卷
// @Tags Exam
// @Summary 交卷（行锁防重，幂等返回已交卷答卷）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.SubmitIn true "交卷请求"
// @Success 200 {object} response.Response{data=model.ExamAnswerSheet} "交卷成功"
// @Router /exam/answer/submit [post]
func (a *ExamApi) SubmitSheet(c *gin.Context) {
	var in service.SubmitIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	sheet, err := service.ExamApp.SubmitSheet(in.SheetId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(sheet, "交卷成功", c)
}

// RunGrade 整卷判分
// @Tags Exam
// @Summary 整卷判分（客观自动 + 主观/实操 LLM，submitted 前置）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.GradeIn true "判分请求"
// @Success 200 {object} response.Response{data=map[string]any} "判分完成"
// @Router /exam/grade/run [post]
func (a *ExamApi) RunGrade(c *gin.Context) {
	var in service.GradeIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	total, err := service.ExamApp.GradeSheet(in.SheetId, in.JudgeModelId, in.MultiPartial)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"total": total}, "判分完成", c)
}

// GetGrades 判分明细
// @Tags Exam
// @Summary 判分明细（按答卷）
// @Security ApiKeyAuth
// @Param sheetId query int true "答卷ID"
// @Success 200 {object} response.Response{data=[]model.ExamGrade} "获取成功"
// @Router /exam/grade/list [get]
func (a *ExamApi) GetGrades(c *gin.Context) {
	sheetId, err := strconv.ParseUint(c.Query("sheetId"), 10, 64)
	if err != nil || sheetId == 0 {
		response.FailWithMessage("sheetId 不能为空", c)
		return
	}
	list, err := service.ExamApp.GetExamGrades(uint(sheetId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// PublishScores 发布成绩
// @Tags Exam
// @Summary 发布成绩（逐卷判分 + 等级判定 + 幂等落档）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.PublishIn true "发布请求"
// @Success 200 {object} response.Response{data=map[string]any} "发布完成"
// @Router /exam/score/publish [post]
func (a *ExamApi) PublishScores(c *gin.Context) {
	var in service.PublishIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	published, err := service.ExamApp.PublishScores(in.ExamId, in.JudgeModelId, in.MultiPartial)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"published": published}, "发布完成", c)
}

// GetScores 成绩列表
// @Tags Exam
// @Summary 成绩单列表（总分降序，分页）
// @Security ApiKeyAuth
// @Param examId query int false "考试ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /exam/score/list [get]
func (a *ExamApi) GetScores(c *gin.Context) {
	examId, _ := strconv.ParseUint(c.Query("examId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ExamApp.GetScores(uint(examId), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// ExportScores 成绩 CSV
// @Tags Exam
// @Summary 成绩 CSV 导出（总分降序）
// @Security ApiKeyAuth
// @Param examId query int true "考试ID"
// @Success 200 {object} response.Response{data=string} "导出成功"
// @Router /exam/score/export [get]
func (a *ExamApi) ExportScores(c *gin.Context) {
	examId, err := strconv.ParseUint(c.Query("examId"), 10, 64)
	if err != nil || examId == 0 {
		response.FailWithMessage("examId 不能为空", c)
		return
	}
	csvText, err := service.ExamApp.ExportScores(uint(examId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(csvText, "导出成功", c)
}

// RegisterNode 节点注册
// @Tags Exam
// @Summary 考试节点注册（upsert 幂等）
// @Accept application/json
// @Param data body service.NodeIn true "节点"
// @Success 200 {object} response.Response{data=model.ExamNode} "注册成功"
// @Router /exam/node/register [post]
func (a *ExamApi) RegisterNode(c *gin.Context) {
	var in service.NodeIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	node, err := service.ExamApp.RegisterNode(in.NodeKey, in.Name, in.Region)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(node, "注册成功", c)
}

// NodeHeartbeat 节点心跳
// @Tags Exam
// @Summary 节点心跳（不存在自动注册，负载回传）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.NodeHeartbeatIn true "心跳"
// @Success 200 {object} response.Response{data=model.ExamNode} "心跳成功"
// @Router /exam/node/heartbeat [post]
func (a *ExamApi) NodeHeartbeat(c *gin.Context) {
	var in service.NodeHeartbeatIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	node, err := service.ExamApp.NodeHeartbeat(in.NodeKey, in.Load)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(node, "心跳成功", c)
}

// GetNodes 节点列表
// @Tags Exam
// @Summary 考试节点列表
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]model.ExamNode} "获取成功"
// @Router /exam/node/list [get]
func (a *ExamApi) GetNodes(c *gin.Context) {
	list, err := service.ExamApp.GetNodes()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// AssignShard 领卷分片
// @Tags Exam
// @Summary 领卷分片分配（region 优先+负载择优，sheetId 幂等）
// @Accept application/json
// @Param data body service.AssignIn true "分配请求"
// @Success 200 {object} response.Response{data=model.ExamShard} "分配成功"
// @Router /exam/node/assign [post]
func (a *ExamApi) AssignShard(c *gin.Context) {
	var in service.AssignIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	shard, err := service.ExamApp.AssignShard(in.ExamId, in.SheetId, in.Region)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(shard, "分配成功", c)
}

// ReportBehavior 行为上报
// @Tags Exam
// @Summary 考试行为上报（切屏/粘贴/复制/退出全屏，白名单）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.BehaviorIn true "行为上报"
// @Success 200 {object} response.Response{msg=string} "上报成功"
// @Router /exam/anticheat/report [post]
func (a *ExamApi) ReportBehavior(c *gin.Context) {
	var in service.BehaviorIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.ExamApp.ReportBehavior(in.SheetId, in.Event, in.Detail); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("上报成功", c)
}

// GetBehaviorStats 行为统计
// @Tags Exam
// @Summary 行为事件计数（按答卷聚合）
// @Security ApiKeyAuth
// @Param sheetId query int true "答卷ID"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /exam/anticheat/stats [get]
func (a *ExamApi) GetBehaviorStats(c *gin.Context) {
	sheetId, err := strconv.ParseUint(c.Query("sheetId"), 10, 64)
	if err != nil || sheetId == 0 {
		response.FailWithMessage("sheetId 不能为空", c)
		return
	}
	stats, err := service.ExamApp.GetBehaviorStats(uint(sheetId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(stats, "获取成功", c)
}

// CheckSimilarity 相似度抽查
// @Tags Exam
// @Summary 主观题答案相似度抽查（同考次已交卷两两比对，可疑对 top N）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.SimCheckIn true "抽查请求"
// @Success 200 {object} response.Response{data=[]service.SuspiciousPair} "抽查完成"
// @Router /exam/anticheat/check [post]
func (a *ExamApi) CheckSimilarity(c *gin.Context) {
	var in service.SimCheckIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	pairs, err := service.ExamApp.CheckSimilarity(in.ExamId, in.SampleSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(pairs, "抽查完成", c)
}

// TerminalRegister 终端注册
// @Tags ExamTerminal
// @Summary 考试终端注册（机器码 upsert，返回节点信息）
// @Accept application/json
// @Param data body service.NodeIn true "终端注册"
// @Success 200 {object} response.Response{data=model.ExamNode} "注册成功"
// @Router /exam/terminal/register [post]
func (a *ExamApi) TerminalRegister(c *gin.Context) {
	var in service.NodeIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	node, err := service.ExamApp.RegisterNode(in.NodeKey, in.Name, in.Region)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(node, "注册成功", c)
}

// TerminalHeartbeat 终端心跳
// @Tags ExamTerminal
// @Description 终端公开面：机器码白名单鉴权（不走管理端 JWT）
// @Summary 终端心跳（机器码白名单）
// @Accept application/json
// @Param data body service.NodeHeartbeatIn true "心跳"
// @Success 200 {object} response.Response{data=model.ExamNode} "心跳成功"
// @Router /exam/terminal/heartbeat [post]
func (a *ExamApi) TerminalHeartbeat(c *gin.Context) {
	var in service.NodeHeartbeatIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if _, err := service.ExamApp.VerifyTerminalKey(in.NodeKey); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	node, err := service.ExamApp.NodeHeartbeat(in.NodeKey, in.Load)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(node, "心跳成功", c)
}

// TerminalExams 可考列表
// @Tags ExamTerminal
// @Description 终端公开面：X-Terminal-Key 白名单鉴权
// @Summary 终端可考列表（机器码白名单；ongoing+时间窗）
// @Param X-Terminal-Key header string true "终端机器码"
// @Success 200 {object} response.Response{data=[]service.TerminalExam} "获取成功"
// @Router /exam/terminal/exams [get]
func (a *ExamApi) TerminalExams(c *gin.Context) {
	if _, err := service.ExamApp.VerifyTerminalKey(c.GetHeader("X-Terminal-Key")); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := service.ExamApp.ListTerminalExams()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// TerminalResume 断线续考
// @Tags ExamTerminal
// @Description 终端公开面：券码即凭证
// @Summary 断线续考（券码幂等恢复题目+已存答案+剩余秒数）
// @Accept application/json
// @Param data body service.ResumeIn true "续考请求"
// @Success 200 {object} response.Response{data=map[string]any} "恢复成功"
// @Router /exam/terminal/resume [post]
func (a *ExamApi) TerminalResume(c *gin.Context) {
	var in service.ResumeIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	qs, sheet, answers, remainSec, err := service.ExamApp.ResumeSession(in.Ticket, in.UserKey)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"questions": qs, "sheet": sheet, "answers": answers, "remainSec": remainSec}, "恢复成功", c)
}

// TerminalSaveAnswers 终端保存答案
// @Tags ExamTerminal
// @Description 终端公开面：券码即凭证（乐观锁）
// @Summary 终端面保存答案（复用乐观锁幂等链）
// @Accept application/json
// @Param data body service.SaveAnswersIn true "保存请求"
// @Success 200 {object} response.Response{data=model.ExamAnswerSheet} "保存成功"
// @Router /exam/terminal/save [post]
func (a *ExamApi) TerminalSaveAnswers(c *gin.Context) {
	var in service.SaveAnswersIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	sheet, err := service.ExamApp.SaveAnswers(in.SheetId, in.Answers, in.Version)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(sheet, "保存成功", c)
}

// TerminalSubmit 终端交卷
// @Tags ExamTerminal
// @Description 终端公开面：券码即凭证（行锁防重）
// @Summary 终端面交卷（复用行锁防重链）
// @Accept application/json
// @Param data body service.SubmitIn true "交卷请求"
// @Success 200 {object} response.Response{data=model.ExamAnswerSheet} "交卷成功"
// @Router /exam/terminal/submit [post]
func (a *ExamApi) TerminalSubmit(c *gin.Context) {
	var in service.SubmitIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	sheet, err := service.ExamApp.SubmitSheet(in.SheetId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(sheet, "交卷成功", c)
}

// RunPractice 实操执行
// @Tags ExamTerminal
// @Description 终端公开面：券码即凭证（过程留痕）
// @Summary 实操题执行（算子/智能体双通道，过程留痕并写回答卷）
// @Accept application/json
// @Param data body service.PracticeIn true "实操请求"
// @Success 200 {object} response.Response{data=model.ExamPracticeLog} "执行完成"
// @Router /exam/terminal/practice [post]
func (a *ExamApi) RunPractice(c *gin.Context) {
	var in service.PracticeIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	log, err := service.ExamApp.RunPractice(in.SheetId, in.QuestionId, in.Channel, in.Target, in.Input)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(log, "执行完成", c)
}

// GetPracticeLogs 实操过程日志
// @Tags ExamTerminal
// @Description 终端公开面：sheetId 即凭证
// @Summary 实操执行历史（对话式作答留痕，倒序 20 条）
// @Param sheetId query int true "答卷ID"
// @Param questionId query int false "实操题ID"
// @Success 200 {object} response.Response{data=[]model.ExamPracticeLog} "获取成功"
// @Router /exam/terminal/practiceLogs [get]
func (a *ExamApi) GetPracticeLogs(c *gin.Context) {
	sheetId, err := strconv.ParseUint(c.Query("sheetId"), 10, 64)
	if err != nil || sheetId == 0 {
		response.FailWithMessage("sheetId 不能为空", c)
		return
	}
	questionId, _ := strconv.ParseUint(c.Query("questionId"), 10, 64)
	list, err := service.ExamApp.GetPracticeLogs(uint(sheetId), uint(questionId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
