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
