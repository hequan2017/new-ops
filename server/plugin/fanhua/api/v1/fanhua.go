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
