// Package v1 ziqiang API：自强-理工科大模型接口层（参数绑定与响应，禁止业务逻辑）
package v1

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"

	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/service"
)

var Api = new(ZiqiangApi)

type ZiqiangApi struct{}

// Ping 插件连通性探测
// @Tags Ziqiang
// @Summary ziqiang 连通性探测
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]any} "探测成功"
// @Router /ziqiang/ping [get]
func (a *ZiqiangApi) Ping(c *gin.Context) {
	response.OkWithDetailed(gin.H{
		"plugin": "ziqiang",
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}, "探测成功", c)
}

// importTextReq 语料导入请求
type importTextReq struct {
	Format  string `json:"format" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// SaveTextCorpus 保存文本语料
// @Tags Ziqiang
// @Summary 新建/更新知识产权文本语料
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.ZqTextCorpus true "语料"
// @Success 200 {object} response.Response{data=model.ZqTextCorpus} "保存成功"
// @Router /ziqiang/textCorpus [post]
func (a *ZiqiangApi) SaveTextCorpus(c *gin.Context) {
	var in model.ZqTextCorpus
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.ZiqiangApp.SaveTextCorpus(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeleteTextCorpus 删除文本语料
// @Tags Ziqiang
// @Summary 删除文本语料（软删除）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "语料ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /ziqiang/textCorpus [delete]
func (a *ZiqiangApi) DeleteTextCorpus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.ZiqiangApp.DeleteTextCorpus(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetTextCorpora 语料列表
// @Tags Ziqiang
// @Summary 文本鉴权语料列表（类型/AIGC/关键字过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param type query string false "专利/开源协议/论文/合同"
// @Param aigc query string false "yes/no"
// @Param keyword query string false "关键字"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /ziqiang/textCorpus/list [get]
func (a *ZiqiangApi) GetTextCorpora(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ZiqiangApp.GetTextCorpora(c.Query("type"), c.Query("aigc"), c.Query("keyword"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// ImportTextCorpora 批量导入
// @Tags Ziqiang
// @Summary 批量导入文本语料（JSONL/CSV 逐行校验+错误行报告）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.importTextReq true "导入请求"
// @Success 200 {object} response.Response{data=map[string]any} "导入完成"
// @Router /ziqiang/textCorpus/import [post]
func (a *ZiqiangApi) ImportTextCorpora(c *gin.Context) {
	var in importTextReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.ZiqiangApp.ImportTextCorpora(in.Format, in.Content)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "导入完成", c)
}

// SaveComparePair 保存对比样本对
// @Tags Ziqiang
// @Summary 保存 AIGC 对比样本对（人创/AIGC 成对校验）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.ZqComparePair true "对比样本对"
// @Success 200 {object} response.Response{data=model.ZqComparePair} "保存成功"
// @Router /ziqiang/comparePair [post]
func (a *ZiqiangApi) SaveComparePair(c *gin.Context) {
	var in model.ZqComparePair
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.ZiqiangApp.SaveComparePair(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// GetComparePairs 对比样本对列表
// @Tags Ziqiang
// @Summary AIGC 对比样本对列表（倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /ziqiang/comparePair/list [get]
func (a *ZiqiangApi) GetComparePairs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ZiqiangApp.GetComparePairs(page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// saveImageCorpusReq 图像语料请求
type saveImageCorpusReq struct {
	Asset      model.ZqImageCorpus `json:"asset"`
	ImageData  string              `json:"imageData"`
}

// SaveImageCorpus 登记图像语料
// @Tags Ziqiang
// @Summary 登记图像鉴权语料（imageData base64 可选，传入即算 aHash）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.saveImageCorpusReq true "图像语料"
// @Success 200 {object} response.Response{data=model.ZqImageCorpus} "登记成功"
// @Router /ziqiang/imageCorpus [post]
func (a *ZiqiangApi) SaveImageCorpus(c *gin.Context) {
	var in saveImageCorpusReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.ZiqiangApp.SaveImageCorpus(&in.Asset, in.ImageData)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "登记成功", c)
}

// GetImageCorpora 图像语料列表
// @Tags Ziqiang
// @Summary 图像鉴权语料列表（图库/关键字过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param gallery query string false "来源图库"
// @Param keyword query string false "名称关键字"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /ziqiang/imageCorpus/list [get]
func (a *ZiqiangApi) GetImageCorpora(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ZiqiangApp.GetImageCorpora(c.Query("gallery"), c.Query("keyword"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// CompareImageHash 哈希比对
// @Tags Ziqiang
// @Summary aHash 汉明距离比对（两哈希 或 与库内语料比对）
// @Security ApiKeyAuth
// @Produce application/json
// @Param hashA query string false "哈希A(16位hex)"
// @Param hashB query string false "哈希B(16位hex)"
// @Param corpusId query int false "库内语料ID（与 hashB 二选一）"
// @Success 200 {object} response.Response{data=service.CompareResult} "比对成功"
// @Router /ziqiang/imageCorpus/compare [get]
func (a *ZiqiangApi) CompareImageHash(c *gin.Context) {
	hashA := c.Query("hashA")
	if hashA == "" {
		response.FailWithMessage("hashA 不能为空", c)
		return
	}
	var (
		out *service.CompareResult
		err error
	)
	if corpusIdStr := c.Query("corpusId"); corpusIdStr != "" {
		corpusId, _ := strconv.ParseUint(corpusIdStr, 10, 64)
		out, err = service.ZiqiangApp.CompareWithCorpus(uint(corpusId), hashA)
	} else {
		out, err = service.ZiqiangApp.CompareHashes(hashA, c.Query("hashB"))
	}
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "比对成功", c)
}

// termCheckReq 术语校验请求
type termCheckReq struct {
	Text string `json:"text" binding:"required"`
}

// dpMaskReq 差分隐私脱敏请求
type dpMaskReq struct {
	Text        string  `json:"text" binding:"required"`
	Sensitivity float64 `json:"sensitivity"`
	Epsilon     float64 `json:"epsilon"`
}

// CheckTerms 法律术语校验
// @Tags Ziqiang
// @Summary 法律术语混淆校验（11 条规则含上下文条件，输出修正建议）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.termCheckReq true "校验请求"
// @Success 200 {object} response.Response{data=[]service.TermIssue} "校验完成"
// @Router /ziqiang/termCheck [post]
func (a *ZiqiangApi) CheckTerms(c *gin.Context) {
	var in termCheckReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(service.CheckTerms(in.Text), "校验完成", c)
}

// DpMaskNumbers 数值差分隐私脱敏
// @Tags Ziqiang
// @Summary 数值 Laplace 噪声脱敏（商业秘密口径，sensitivity/epsilon 可配）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.dpMaskReq true "脱敏请求"
// @Success 200 {object} response.Response{data=map[string]any} "脱敏完成"
// @Router /ziqiang/dpMask [post]
func (a *ZiqiangApi) DpMaskNumbers(c *gin.Context) {
	var in dpMaskReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if in.Sensitivity <= 0 {
		in.Sensitivity = 1
	}
	if in.Epsilon <= 0 {
		in.Epsilon = 1
	}
	masked, err := service.DpMaskNumbers(in.Text, in.Sensitivity, in.Epsilon, rand.New(rand.NewSource(time.Now().UnixNano())))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"masked": masked}, "脱敏完成", c)
}

// legalTokenizeReq 法律分词请求
type legalTokenizeReq struct {
	Text string `json:"text" binding:"required"`
}

// plagiarismReq 抄袭识别请求
type plagiarismReq struct {
	Source    string  `json:"source" binding:"required"`
	Suspicious string `json:"suspicious" binding:"required"`
	Window    int     `json:"window"`
	Threshold float64 `json:"threshold"`
}

// LegalTokenize 法律分词
// @Tags Ziqiang
// @Summary 法律文本分词（50 复合术语词典最大正向匹配，OOV 单字）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.legalTokenizeReq true "分词请求"
// @Success 200 {object} response.Response{data=[]service.LegalToken} "分词成功"
// @Router /ziqiang/legalTokenize [post]
func (a *ZiqiangApi) LegalTokenize(c *gin.Context) {
	var in legalTokenizeReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(service.LegalSegment(in.Text), "分词成功", c)
}

// SlidingPlagiarism 隐性抄袭识别
// @Tags Ziqiang
// @Summary 滑窗 bigram 隐性抄袭片段识别（窗口/阈值可配）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.plagiarismReq true "识别请求"
// @Success 200 {object} response.Response{data=[]service.PlagiarismHit} "识别完成"
// @Router /ziqiang/plagiarismCheck [post]
func (a *ZiqiangApi) SlidingPlagiarism(c *gin.Context) {
	var in plagiarismReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if in.Window == 0 {
		in.Window = 40
	}
	if in.Threshold == 0 {
		in.Threshold = 0.8
	}
	hits, err := service.SlidingPlagiarism(in.Source, in.Suspicious, in.Window, in.Threshold)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(hits, "识别完成", c)
}

// SaveGraphEntity 保存实体
// @Tags Ziqiang
// @Summary 新建/更新知识产权实体（法律/技术/权利主体）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.ZqEntity true "实体"
// @Success 200 {object} response.Response{data=model.ZqEntity} "保存成功"
// @Router /ziqiang/entity [post]
func (a *ZiqiangApi) SaveGraphEntity(c *gin.Context) {
	var in model.ZqEntity
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.ZiqiangApp.SaveEntity(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeleteGraphEntity 删除实体
// @Tags Ziqiang
// @Summary 删除知识产权实体（软删除）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "实体ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /ziqiang/entity [delete]
func (a *ZiqiangApi) DeleteGraphEntity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.ZiqiangApp.DeleteEntity(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetGraphEntities 实体列表
// @Tags Ziqiang
// @Summary 知识产权实体列表（分类过滤）
// @Security ApiKeyAuth
// @Produce application/json
// @Param category query string false "法律/技术/权利主体"
// @Success 200 {object} response.Response{data=[]model.ZqEntity} "获取成功"
// @Router /ziqiang/entity/list [get]
func (a *ZiqiangApi) GetGraphEntities(c *gin.Context) {
	list, err := service.ZiqiangApp.GetEntities(c.Query("category"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveGraphEdge 保存关系边
// @Tags Ziqiang
// @Summary 新建实体关系边（禁自环）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.ZqEdge true "关系边"
// @Success 200 {object} response.Response{data=model.ZqEdge} "保存成功"
// @Router /ziqiang/edge [post]
func (a *ZiqiangApi) SaveGraphEdge(c *gin.Context) {
	var in model.ZqEdge
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.ZiqiangApp.SaveEdge(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeleteGraphEdge 删除关系边
// @Tags Ziqiang
// @Summary 删除关系边
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "边ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /ziqiang/edge [delete]
func (a *ZiqiangApi) DeleteGraphEdge(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.ZiqiangApp.DeleteEdge(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetGraph 全图
// @Tags Ziqiang
// @Summary 知识产权图谱全图（节点+边）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /ziqiang/graph [get]
func (a *ZiqiangApi) GetGraph(c *gin.Context) {
	out, err := service.ZiqiangApp.GetGraph()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "获取成功", c)
}

// GraphBFS 多跳推理
// @Tags Ziqiang
// @Summary 图谱 BFS 多跳推理（最短路径，hop 1-6）
// @Security ApiKeyAuth
// @Produce application/json
// @Param startId query int true "起点实体ID"
// @Param maxHops query int false "最大跳数(默认2)"
// @Success 200 {object} response.Response{data=[]service.graphPath} "推理成功"
// @Router /ziqiang/graph/bfs [get]
func (a *ZiqiangApi) GraphBFS(c *gin.Context) {
	startId, _ := strconv.ParseUint(c.Query("startId"), 10, 64)
	maxHops, _ := strconv.Atoi(c.DefaultQuery("maxHops", "2"))
	paths, err := service.ZiqiangApp.GraphBFS(uint(startId), maxHops)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(paths, "推理成功", c)
}

// SaveRegulation 保存法规
// @Tags Ziqiang
// @Summary 新建/更新法规条目（中美欧三地）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.ZqRegulation true "法规"
// @Success 200 {object} response.Response{data=model.ZqRegulation} "保存成功"
// @Router /ziqiang/regulation [post]
func (a *ZiqiangApi) SaveRegulation(c *gin.Context) {
	var in model.ZqRegulation
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.ZiqiangApp.SaveRegulation(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeleteRegulation 删除法规
// @Tags Ziqiang
// @Summary 删除法规条目（软删除）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "法规ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /ziqiang/regulation [delete]
func (a *ZiqiangApi) DeleteRegulation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.ZiqiangApp.DeleteRegulation(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetRegulations 法规列表
// @Tags Ziqiang
// @Summary 法规条目列表（辖区/分类过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param jurisdiction query string false "中国/美国/欧盟"
// @Param category query string false "分类"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /ziqiang/regulation/list [get]
func (a *ZiqiangApi) GetRegulations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ZiqiangApp.GetRegulations(c.Query("jurisdiction"), c.Query("category"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// SearchRegulations 法规检索
// @Tags Ziqiang
// @Summary 法规 bigram 检索（辖区过滤，命中降序 top-k）
// @Security ApiKeyAuth
// @Produce application/json
// @Param query query string true "检索词"
// @Param jurisdiction query string false "中国/美国/欧盟"
// @Param topK query int false "返回条数"
// @Success 200 {object} response.Response{data=[]service.regHit} "检索成功"
// @Router /ziqiang/regulation/search [get]
func (a *ZiqiangApi) SearchRegulations(c *gin.Context) {
	topK, _ := strconv.Atoi(c.DefaultQuery("topK", "5"))
	hits, err := service.ZiqiangApp.SearchRegulations(c.Query("query"), c.Query("jurisdiction"), topK)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(hits, "检索成功", c)
}

// SaveCase 保存判例
// @Tags Ziqiang
// @Summary 新建判例条目
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.ZqCase true "判例"
// @Success 200 {object} response.Response{data=model.ZqCase} "保存成功"
// @Router /ziqiang/case [post]
func (a *ZiqiangApi) SaveCase(c *gin.Context) {
	var in model.ZqCase
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.ZiqiangApp.SaveCase(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// SearchCases 判例检索
// @Tags Ziqiang
// @Summary 判例 bigram 检索（辖区过滤，命中降序 top-k，口径同法规检索）
// @Security ApiKeyAuth
// @Produce application/json
// @Param query query string true "检索词"
// @Param jurisdiction query string false "中国/美国/欧盟"
// @Param topK query int false "返回条数"
// @Success 200 {object} response.Response{data=[]service.CaseHit} "检索成功"
// @Router /ziqiang/case/search [get]
func (a *ZiqiangApi) SearchCases(c *gin.Context) {
	topK, _ := strconv.Atoi(c.DefaultQuery("topK", "5"))
	hits, err := service.ZiqiangApp.SearchCases(c.Query("query"), c.Query("jurisdiction"), topK)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(hits, "检索成功", c)
}

// GetCases 判例列表
// @Tags Ziqiang
// @Summary 判例条目列表（辖区过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param jurisdiction query string false "中国/美国/欧盟"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /ziqiang/case/list [get]
func (a *ZiqiangApi) GetCases(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ZiqiangApp.GetCases(c.Query("jurisdiction"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// evidenceReq 存证追加请求
type evidenceReq struct {
	Payload string `json:"payload" binding:"required"`
	BizKey  string `json:"bizKey"`
}

// AppendEvidence 追加存证
// @Tags Ziqiang
// @Summary 区块链存证追加（SHA256 链式：dataHash/prevHash/height/timestamp）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.evidenceReq true "存证请求"
// @Success 200 {object} response.Response{data=model.ZqEvidence} "存证成功"
// @Router /ziqiang/evidence [post]
func (a *ZiqiangApi) AppendEvidence(c *gin.Context) {
	var in evidenceReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.ZiqiangApp.AppendEvidence(in.Payload, in.BizKey)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "存证成功", c)
}

// VerifyEvidenceChain 链完整性校验
// @Tags Ziqiang
// @Summary 区块链完整性校验（块哈希/前链指针/原文哈希三重重算，定位首个篡改块）
// @Security ApiKeyAuth
// @Produce application/json
// @Param bizKey query string false "业务号（传则局部核验该业务全部块）"
// @Success 200 {object} response.Response{data=map[string]any} "校验完成"
// @Router /ziqiang/evidence/verify [get]
func (a *ZiqiangApi) VerifyEvidenceChain(c *gin.Context) {
	var (
		out *service.ChainReport
		err error
	)
	if bizKey := c.Query("bizKey"); bizKey != "" {
		out, err = service.ZiqiangApp.VerifyBizChain(bizKey)
	} else {
		out, err = service.ZiqiangApp.VerifyChain()
	}
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "校验完成", c)
}

// ChainHead 链头摘要
// @Tags Ziqiang
// @Summary 区块链链头（最新块高/块哈希/总量/时间戳，空链为创世口径）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /ziqiang/evidence/head [get]
func (a *ZiqiangApi) ChainHead(c *gin.Context) {
	out, err := service.ZiqiangApp.ChainHead()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "获取成功", c)
}

// GetEvidenceList 存证列表
// @Tags Ziqiang
// @Summary 存证块列表（业务号过滤，块高升序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param bizKey query string false "业务关联号"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /ziqiang/evidence/list [get]
func (a *ZiqiangApi) GetEvidenceList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.ZiqiangApp.GetEvidenceList(c.Query("bizKey"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetEvidence 存证详情
// @Tags Ziqiang
// @Summary 存证块详情（按 ID）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "存证ID"
// @Success 200 {object} response.Response{data=model.ZqEvidence} "获取成功"
// @Router /ziqiang/evidence [get]
func (a *ZiqiangApi) GetEvidence(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	out, err := service.ZiqiangApp.GetEvidence(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "获取成功", c)
}

// SavePrompt 保存提示词模板
// @Tags Ziqiang
// @Summary 新建/更新侵权检测提示词模板（vars 自动解析，场景白名单）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.ZqPromptTemplate true "模板"
// @Success 200 {object} response.Response{data=model.ZqPromptTemplate} "保存成功"
// @Router /ziqiang/prompt [post]
func (a *ZiqiangApi) SavePrompt(c *gin.Context) {
	var in model.ZqPromptTemplate
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.ZiqiangApp.SavePromptTemplate(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "保存成功", c)
}

// DeletePrompt 删除提示词模板
// @Tags Ziqiang
// @Summary 删除提示词模板（软删除）
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "模板ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /ziqiang/prompt [delete]
func (a *ZiqiangApi) DeletePrompt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.ZiqiangApp.DeletePromptTemplate(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetPrompts 提示词模板列表
// @Tags Ziqiang
// @Summary 提示词模板列表（场景过滤，场景+步号升序）
// @Security ApiKeyAuth
// @Produce application/json
// @Param scene query string false "多模态双验证/引导式对话"
// @Success 200 {object} response.Response{data=[]model.ZqPromptTemplate} "获取成功"
// @Router /ziqiang/prompt/list [get]
func (a *ZiqiangApi) GetPrompts(c *gin.Context) {
	list, err := service.ZiqiangApp.GetPromptTemplates(c.Query("scene"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// renderPromptReq 渲染请求
type renderPromptReq struct {
	Name string            `json:"name" binding:"required"`
	Vars map[string]string `json:"vars"`
}

// RenderPrompt 渲染提示词
// @Tags Ziqiang
// @Summary 按模板名渲染（缺失占位变量报错）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.renderPromptReq true "渲染请求"
// @Success 200 {object} response.Response{data=map[string]any} "渲染成功"
// @Router /ziqiang/prompt/render [post]
func (a *ZiqiangApi) RenderPrompt(c *gin.Context) {
	var in renderPromptReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.ZiqiangApp.RenderPromptByName(in.Name, in.Vars)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "渲染成功", c)
}

// qaPatentReq 知识产权问答请求
type qaPatentReq struct {
	Text      string `json:"text" binding:"required"`
	ImageHash string `json:"imageHash"`
	TopK      int    `json:"topK"`
}

// QaPatentSearch 知识产权问答
// @Tags Ziqiang
// @Summary 知识产权问答（文本+图像混合输入 → 相似专利检索 bigram+aHash 融合排序 → 风险评估报告 JSON）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body v1.qaPatentReq true "问答请求"
// @Success 200 {object} response.Response{data=service.qaReport} "评估完成"
// @Router /ziqiang/qa/patent [post]
func (a *ZiqiangApi) QaPatentSearch(c *gin.Context) {
	var in qaPatentReq
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	out, err := service.ZiqiangApp.QaPatentSearch(in.Text, in.ImageHash, in.TopK)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(out, "评估完成", c)
}
