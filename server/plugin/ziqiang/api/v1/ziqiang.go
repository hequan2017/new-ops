// Package v1 ziqiang API：自强-理工科大模型接口层（参数绑定与响应，禁止业务逻辑）
package v1

import (
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
