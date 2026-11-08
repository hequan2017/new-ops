// Package v1 fanhua API：繁花-文科大模型接口层（参数绑定与响应，禁止业务逻辑）
package v1

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/service"
)

var Api = new(FanhuaApi)

type FanhuaApi struct{}

// importCorpusReq 语料导入请求
type importCorpusReq struct {
	Format  string `json:"format" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// Ping 插件连通性探测
// @Tags Fanhua
// @Summary fanhua 连通性探测
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]any} "探测成功"
// @Router /fanhua/ping [get]
func (a *FanhuaApi) Ping(c *gin.Context) {
	response.OkWithDetailed(gin.H{
		"plugin": "fanhua",
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}, "探测成功", c)
}

// SaveCorpus 保存语料
// @Tags Fanhua
// @Summary 新建/更新沪语语料
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.FanhuaCorpus true "语料"
// @Success 200 {object} response.Response{data=model.FanhuaCorpus} "保存成功"
// @Router /fanhua/corpus [post]
func (a *FanhuaApi) SaveCorpus(c *gin.Context) {
	var in model.FanhuaCorpus
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.SaveCorpus(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeleteCorpus 删除语料
// @Tags Fanhua
// @Summary 删除沪语语料（软删除）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "语料ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /fanhua/corpus [delete]
func (a *FanhuaApi) DeleteCorpus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.FanhuaApp.DeleteCorpus(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetCorpora 语料列表
// @Tags Fanhua
// @Summary 沪语语料列表（来源/语种/关键字过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param source query string false "新闻/社交/文学/对照/舆情/城市"
// @Param lang query string false "沪语/普通话/英语/混合"
// @Param keyword query string false "标题/正文关键字"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /fanhua/corpus/list [get]
func (a *FanhuaApi) GetCorpora(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.FanhuaApp.GetCorpora(c.Query("source"), c.Query("lang"), c.Query("keyword"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// ImportCorpora 批量导入语料
// @Tags Fanhua
// @Summary 批量导入语料（JSONL/CSV 逐行校验，错误行报告）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.importCorpusReq true "导入请求"
// @Success 200 {object} response.Response{data=map[string]any} "导入完成"
// @Router /fanhua/corpus/import [post]
func (a *FanhuaApi) ImportCorpora(c *gin.Context) {
	var in importCorpusReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.ImportCorpora(in.Format, in.Content)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "导入完成", c)
}

// SaveParallel 保存平行语料
// @Tags Fanhua
// @Summary 新建/更新沪-普-英平行语料
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.FanhuaParallel true "平行语料"
// @Success 200 {object} response.Response{data=model.FanhuaParallel} "保存成功"
// @Router /fanhua/parallel [post]
func (a *FanhuaApi) SaveParallel(c *gin.Context) {
	var in model.FanhuaParallel
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.SaveParallel(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeleteParallel 删除平行语料
// @Tags Fanhua
// @Summary 删除平行语料（软删除）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "语料ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /fanhua/parallel [delete]
func (a *FanhuaApi) DeleteParallel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.FanhuaApp.DeleteParallel(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetParallels 平行语料列表
// @Tags Fanhua
// @Summary 沪-普-英平行语料列表（关键字过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param keyword query string false "关键字"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /fanhua/parallel/list [get]
func (a *FanhuaApi) GetParallels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.FanhuaApp.GetParallels(c.Query("keyword"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// SearchParallel 对照检索
// @Tags Fanhua
// @Summary 普通话→沪语对照检索（编辑相似度 top-k）
// @Security ApiKeyAuth
// @Produce application/json
// @Param query query string true "检索语句"
// @Param topK query int false "返回条数(默认5)"
// @Success 200 {object} response.Response{data=[]service.paraHit} "检索成功"
// @Router /fanhua/parallel/search [get]
func (a *FanhuaApi) SearchParallel(c *gin.Context) {
	topK, _ := strconv.Atoi(c.DefaultQuery("topK", "5"))
	hits, err := service.FanhuaApp.SearchParallel(c.Query("query"), topK)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(hits, "检索成功", c)
}

// cleanDryReq 清洗试运行请求
type cleanDryReq struct {
	Text   string `json:"text" binding:"required"`
	Dedupe bool   `json:"dedupe"`
}

// cleanIdsReq 语料清洗请求
type cleanIdsReq struct {
	IDs    []uint `json:"ids" binding:"required"`
	Dedupe bool   `json:"dedupe"`
}

// CleanDry 清洗试运行
// @Tags Fanhua
// @Summary 单文本清洗试运行（零宽/脱敏/乱码/重复行，不落库）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.cleanDryReq true "清洗请求"
// @Success 200 {object} response.Response{data=service.CleanReport} "清洗成功"
// @Router /fanhua/clean/dry [post]
func (a *FanhuaApi) CleanDry(c *gin.Context) {
	var in cleanDryReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	text, stats := service.CleanText(in.Text, in.Dedupe)
	response.OkWithDetailed(service.CleanReport{Text: text, Stats: stats}, "清洗成功", c)
}

// CleanCorpora 语料批量清洗
// @Tags Fanhua
// @Summary 语料库批量清洗（按 ID 清洗正文并回写，逐条报告）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.cleanIdsReq true "清洗请求"
// @Success 200 {object} response.Response{data=[]service.CleanReport} "清洗完成"
// @Router /fanhua/clean [post]
func (a *FanhuaApi) CleanCorpora(c *gin.Context) {
	var in cleanIdsReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	reports, err := service.FanhuaApp.CleanCorpora(in.IDs, in.Dedupe)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(reports, "清洗完成", c)
}

// tokenizeReq 分词请求
type tokenizeReq struct {
	Text string `json:"text" binding:"required"`
}

// Tokenize 沪语分词
// @Tags Fanhua
// @Summary 沪语分词（专用词典最大正向匹配+OOV bigram 兜底，附词频降序）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.tokenizeReq true "分词请求"
// @Success 200 {object} response.Response{data=map[string]any} "分词成功"
// @Router /fanhua/tokenize [post]
func (a *FanhuaApi) Tokenize(c *gin.Context) {
	var in tokenizeReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{
		"tokens": service.Segment(in.Text),
		"freq":   service.WordFreq(in.Text),
	}, "分词成功", c)
}

// annotateReq 实体标注请求
type annotateReq struct {
	Text     string `json:"text" binding:"required"`
	CorpusId uint   `json:"corpusId"`
}

// SaveEntity 保存实体
// @Tags Fanhua
// @Summary 新建/更新海派文化实体
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.FanhuaEntity true "实体"
// @Success 200 {object} response.Response{data=model.FanhuaEntity} "保存成功"
// @Router /fanhua/entity [post]
func (a *FanhuaApi) SaveEntity(c *gin.Context) {
	var in model.FanhuaEntity
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.SaveEntity(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeleteEntity 删除实体
// @Tags Fanhua
// @Summary 删除海派文化实体（软删除）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "实体ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /fanhua/entity [delete]
func (a *FanhuaApi) DeleteEntity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.FanhuaApp.DeleteEntity(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetEntities 实体列表
// @Tags Fanhua
// @Summary 海派文化实体列表（分类过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param category query string false "地域/文化符号/地标"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /fanhua/entity/list [get]
func (a *FanhuaApi) GetEntities(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.FanhuaApp.GetEntities(c.Query("category"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// AnnotateText 实体标注
// @Tags Fanhua
// @Summary 词典匹配实体标注（词条+别名计数降序，结果落库）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.annotateReq true "标注请求"
// @Success 200 {object} response.Response{data=map[string]any} "标注成功"
// @Router /fanhua/entity/annotate [post]
func (a *FanhuaApi) AnnotateText(c *gin.Context) {
	var in annotateReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.AnnotateText(in.Text, in.CorpusId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "标注成功", c)
}

// GetEntityHits 标注记录列表
// @Tags Fanhua
// @Summary 实体标注记录列表（倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /fanhua/entity/hits [get]
func (a *FanhuaApi) GetEntityHits(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.FanhuaApp.GetEntityHits(page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}
