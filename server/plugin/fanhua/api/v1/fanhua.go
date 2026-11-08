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

// buildLabelReq 标注构建请求
type buildLabelReq struct {
	TrainRatio float64 `json:"trainRatio"`
	Limit      int     `json:"limit"`
}

// updateLabelReq 人工改标请求
type updateLabelReq struct {
	ID        uint   `json:"id" binding:"required"`
	Sentiment string `json:"sentiment" binding:"required"`
	Split     string `json:"split" binding:"required"`
}

// BuildTextLabels 构建标注
// @Tags Fanhua
// @Summary 从语料库构建文本标注（train/test 确定性划分+情感规则预标）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.buildLabelReq true "构建请求"
// @Success 200 {object} response.Response{data=map[string]any} "构建完成"
// @Router /fanhua/label/build [post]
func (a *FanhuaApi) BuildTextLabels(c *gin.Context) {
	var in buildLabelReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if in.TrainRatio == 0 {
		in.TrainRatio = 0.8
	}
	out, err := service.FanhuaApp.BuildTextLabels(in.TrainRatio, in.Limit)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "构建完成", c)
}

// UpdateLabel 人工改标
// @Tags Fanhua
// @Summary 人工改标（情感/划分覆盖，置 Labeled）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.updateLabelReq true "改标请求"
// @Success 200 {object} response.Response{data=model.FanhuaTextLabel} "改标成功"
// @Router /fanhua/label [put]
func (a *FanhuaApi) UpdateLabel(c *gin.Context) {
	var in updateLabelReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.UpdateLabel(in.ID, in.Sentiment, in.Split)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "改标成功", c)
}

// GetTextLabels 标注列表
// @Tags Fanhua
// @Summary 文本标注列表（划分/情感过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param split query string false "train/test/none"
// @Param sentiment query string false "正面/中立/负面"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /fanhua/label/list [get]
func (a *FanhuaApi) GetTextLabels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.FanhuaApp.GetTextLabels(c.Query("split"), c.Query("sentiment"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// LabelStats 标注分布统计
// @Tags Fanhua
// @Summary 标注分布统计（划分×情感聚合）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /fanhua/label/stats [get]
func (a *FanhuaApi) LabelStats(c *gin.Context) {
	out, err := service.FanhuaApp.LabelStats()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "获取成功", c)
}

// labelImageReq 图片标签请求
type labelImageReq struct {
	AssetId  uint   `json:"assetId" binding:"required"`
	TagsJSON string `json:"tagsJSON" binding:"required"`
}

// SaveImageAsset 登记图片资产
// @Tags Fanhua
// @Summary 登记图片资产（路径/MIME/train-test 用途）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.FanhuaImageAsset true "图片资产"
// @Success 200 {object} response.Response{data=model.FanhuaImageAsset} "登记成功"
// @Router /fanhua/imageAsset [post]
func (a *FanhuaApi) SaveImageAsset(c *gin.Context) {
	var in model.FanhuaImageAsset
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.SaveImageAsset(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "登记成功", c)
}

// GetImageAssets 图片资产列表
// @Tags Fanhua
// @Summary 图片资产列表（用途/关键字过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param split query string false "train/test/none"
// @Param keyword query string false "名称关键字"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /fanhua/imageAsset/list [get]
func (a *FanhuaApi) GetImageAssets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.FanhuaApp.GetImageAssets(c.Query("split"), c.Query("keyword"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// LabelImage 写图片标签
// @Tags Fanhua
// @Summary 图片标签 upsert（一图一条，写即置已标注）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.labelImageReq true "标签请求"
// @Success 200 {object} response.Response{data=model.FanhuaImageLabel} "标注成功"
// @Router /fanhua/imageAsset/label [post]
func (a *FanhuaApi) LabelImage(c *gin.Context) {
	var in labelImageReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.LabelImage(in.AssetId, in.TagsJSON)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "标注成功", c)
}

// GetImageLabels 图片标签列表
// @Tags Fanhua
// @Summary 图片标签列表（标注状态过滤，倒序分页，含资产信息）
// @Security ApiKeyAuth
// @Produce application/json
// @Param labeled query string false "yes/no"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /fanhua/imageLabel/list [get]
func (a *FanhuaApi) GetImageLabels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.FanhuaApp.GetImageLabels(c.Query("labeled"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// searchCultureReq 海派文化组合检索请求
type searchCultureReq struct {
	Query       string `json:"query" binding:"required"`
	Category    string `json:"category"`
	SpeakerType string `json:"speakerType"`
	TopK        int    `json:"topK"`
}

// voiceIdReq 语音条目删除请求
type voiceDeleteReq struct{}

// SaveVoiceEntry 保存语音条目
// @Tags Fanhua
// @Summary 新建/更新海派文化语音条目
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.FanhuaVoiceEntry true "语音条目"
// @Success 200 {object} response.Response{data=model.FanhuaVoiceEntry} "保存成功"
// @Router /fanhua/voice [post]
func (a *FanhuaApi) SaveVoiceEntry(c *gin.Context) {
	var in model.FanhuaVoiceEntry
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.SaveVoiceEntry(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeleteVoiceEntry 删除语音条目
// @Tags Fanhua
// @Summary 删除语音条目（软删除）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "条目ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /fanhua/voice [delete]
func (a *FanhuaApi) DeleteVoiceEntry(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.FanhuaApp.DeleteVoiceEntry(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetVoiceEntries 语音条目列表
// @Tags Fanhua
// @Summary 海派文化语音条目列表（分类/声部过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param category query string false "沪剧台词/弄堂俚语/经典朗读"
// @Param speakerType query string false "男声/女声/童声"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /fanhua/voice/list [get]
func (a *FanhuaApi) GetVoiceEntries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.FanhuaApp.GetVoiceEntries(c.Query("category"), c.Query("speakerType"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// SearchCulture 海派文化组合检索
// @Tags Fanhua
// @Summary 组合检索（语音条目分词命中打分+平行语料相似度双路）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.searchCultureReq true "检索请求"
// @Success 200 {object} response.Response{data=map[string]any} "检索成功"
// @Router /fanhua/voice/search [post]
func (a *FanhuaApi) SearchCulture(c *gin.Context) {
	var in searchCultureReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.SearchCulture(in.Query, in.Category, in.SpeakerType, in.TopK)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "检索成功", c)
}

// healthReq 健康度评估请求
type healthReq struct {
	Text string `json:"text" binding:"required"`
}

// GetWords 敏感词列表
// @Tags Fanhua
// @Summary 敏感词列表（动作过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param action query string false "block/replace"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /fanhua/word/list [get]
func (a *FanhuaApi) GetWords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.FanhuaApp.GetWords(c.Query("action"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// SaveWord 保存敏感词
// @Tags Fanhua
// @Summary 新建/更新敏感词（block/replace+权重）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.FanhuaSensitiveWord true "敏感词"
// @Success 200 {object} response.Response{data=model.FanhuaSensitiveWord} "保存成功"
// @Router /fanhua/word [post]
func (a *FanhuaApi) SaveWord(c *gin.Context) {
	var in model.FanhuaSensitiveWord
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.SaveWord(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeleteWord 删除敏感词
// @Tags Fanhua
// @Summary 删除敏感词（软删除）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "词ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /fanhua/word [delete]
func (a *FanhuaApi) DeleteWord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.FanhuaApp.DeleteWord(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetHealth 健康度评估
// @Tags Fanhua
// @Summary 文本价值观健康度评估（加权扣减+情感三分类+替换输出）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.healthReq true "评估请求"
// @Success 200 {object} response.Response{data=service.healthResult} "评估成功"
// @Router /fanhua/values/health [post]
func (a *FanhuaApi) GetHealth(c *gin.Context) {
	var in healthReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.GetHealth(in.Text)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "评估成功", c)
}

// GetTemplates 合规模板列表
// @Tags Fanhua
// @Summary 合规表达模板列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.FanhuaComplianceTemplate} "获取成功"
// @Router /fanhua/template/list [get]
func (a *FanhuaApi) GetTemplates(c *gin.Context) {
	list, err := service.FanhuaApp.GetTemplates()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// GetConflictCases 冲突案例列表
// @Tags Fanhua
// @Summary 多语言价值观冲突案例列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param lang query string false "中/英/多语"
// @Success 200 {object} response.Response{data=[]model.FanhuaConflictCase} "获取成功"
// @Router /fanhua/conflict/list [get]
func (a *FanhuaApi) GetConflictCases(c *gin.Context) {
	list, err := service.FanhuaApp.GetConflictCases(c.Query("lang"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveIndicator 保存治理指标
// @Tags Fanhua
// @Summary 新建/更新城市治理指标（父子树，含视觉维度）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.FanhuaGovIndicator true "指标"
// @Success 200 {object} response.Response{data=model.FanhuaGovIndicator} "保存成功"
// @Router /fanhua/govIndicator [post]
func (a *FanhuaApi) SaveIndicator(c *gin.Context) {
	var in model.FanhuaGovIndicator
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.SaveIndicator(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeleteIndicator 删除治理指标
// @Tags Fanhua
// @Summary 删除治理指标（软删除）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "指标ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /fanhua/govIndicator [delete]
func (a *FanhuaApi) DeleteIndicator(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.FanhuaApp.DeleteIndicator(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetIndicatorTree 指标树
// @Tags Fanhua
// @Summary 城市治理指标树
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]service.indicatorNode} "获取成功"
// @Router /fanhua/govIndicator/list [get]
func (a *FanhuaApi) GetIndicatorTree(c *gin.Context) {
	roots, err := service.FanhuaApp.GetIndicatorTree()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(roots, "获取成功", c)
}

// SaveDataPoint 写数据点
// @Tags Fanhua
// @Summary 写入区域指标数据点（问卷/POI/传感聚合值 0-100）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.FanhuaGovDataPoint true "数据点"
// @Success 200 {object} response.Response{data=model.FanhuaGovDataPoint} "写入成功"
// @Router /fanhua/govData [post]
func (a *FanhuaApi) SaveDataPoint(c *gin.Context) {
	var in model.FanhuaGovDataPoint
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.SaveDataPoint(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "写入成功", c)
}

// GetDataPoints 数据点列表
// @Tags Fanhua
// @Summary 区域指标数据点列表（区域/期间过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param region query string false "区域"
// @Param period query string false "期间 YYYY-MM"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /fanhua/govData/list [get]
func (a *FanhuaApi) GetDataPoints(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.FanhuaApp.GetDataPoints(c.Query("region"), c.Query("period"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetRegionProfile 社区画像
// @Tags Fanhua
// @Summary 社区画像（指标树加权得分+数据点计数）
// @Security ApiKeyAuth
// @Produce application/json
// @Param region query string true "区域"
// @Param period query string false "期间 YYYY-MM"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /fanhua/govProfile [get]
func (a *FanhuaApi) GetRegionProfile(c *gin.Context) {
	out, err := service.FanhuaApp.GetRegionProfile(c.Query("region"), c.Query("period"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "获取成功", c)
}

// renderPromptReq 渲染请求
type renderPromptReq struct {
	Name string            `json:"name" binding:"required"`
	Vars map[string]string `json:"vars"`
}

// SavePromptTemplate 保存提示词模板
// @Tags Fanhua
// @Summary 新建/更新提示词模板（vars 自动解析）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.FanhuaPromptTemplate true "模板"
// @Success 200 {object} response.Response{data=model.FanhuaPromptTemplate} "保存成功"
// @Router /fanhua/prompt [post]
func (a *FanhuaApi) SavePromptTemplate(c *gin.Context) {
	var in model.FanhuaPromptTemplate
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.SavePromptTemplate(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeletePromptTemplate 删除提示词模板
// @Tags Fanhua
// @Summary 删除提示词模板（软删除）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "模板ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /fanhua/prompt [delete]
func (a *FanhuaApi) DeletePromptTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.FanhuaApp.DeletePromptTemplate(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetPromptTemplates 模板列表
// @Tags Fanhua
// @Summary 提示词模板列表（场景过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param scene query string false "沪语场景/价值观审核/通用"
// @Success 200 {object} response.Response{data=[]model.FanhuaPromptTemplate} "获取成功"
// @Router /fanhua/prompt/list [get]
func (a *FanhuaApi) GetPromptTemplates(c *gin.Context) {
	list, err := service.FanhuaApp.GetPromptTemplates(c.Query("scene"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// RenderPrompt 模板渲染
// @Tags Fanhua
// @Summary 按模板名渲染（{{var}} 填充，缺失报错）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.renderPromptReq true "渲染请求"
// @Success 200 {object} response.Response{data=map[string]any} "渲染成功"
// @Router /fanhua/prompt/render [post]
func (a *FanhuaApi) RenderPrompt(c *gin.Context) {
	var in renderPromptReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.FanhuaApp.RenderPromptByName(in.Name, in.Vars)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "渲染成功", c)
}
