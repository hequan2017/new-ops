// Package v1 llmops API：大模型应用开发平台接口层（参数绑定与响应，禁止业务逻辑）
package v1

import (
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hequan2017/new-ops/server/model/common/response"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
	"github.com/hequan2017/new-ops/server/plugin/llmops/service"
)

// ImportQaMaxBytes 问答对文件上传上限（10MB）
const ImportQaMaxBytes int64 = 10 << 20

var Api = new(LlmopsApi)

type LlmopsApi struct{}

// webImportReq 网页采集请求
type webImportReq struct {
	URL  string `json:"url" binding:"required"`
	Name string `json:"name"`
	Tags string `json:"tags"`
}

// categorizeReq 资产挂分类/打标签请求
type categorizeReq struct {
	AssetId    uint   `json:"assetId" binding:"required"`
	CategoryId uint   `json:"categoryId"`
	Tags       string `json:"tags"`
}

// Ping 插件连通性探测
// @Tags Llmops
// @Summary llmops 连通性探测
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]any} "探测成功"
// @Router /llmops/ping [get]
func (a *LlmopsApi) Ping(c *gin.Context) {
	response.OkWithDetailed(gin.H{
		"plugin": "llmops",
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}, "探测成功", c)
}

// GetAssets 数据资产列表
// @Tags Llmops
// @Summary 数据资产列表（类型/状态/关键字过滤，白名单校验）
// @Security ApiKeyAuth
// @Produce application/json
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param dataType query string false "数据类型(doc/image/audio/video/webpage/qa/structured/semistructured)"
// @Param status query string false "状态(draft/parsed/cataloged/offline)"
// @Param keyword query string false "名称/标签关键字"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /llmops/asset/list [get]
func (a *LlmopsApi) GetAssets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	list, total, err := service.LlmopsService.GetAssetList(page, pageSize,
		c.Query("dataType"), c.Query("status"), c.Query("keyword"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetAsset 数据资产详情
// @Tags Llmops
// @Summary 数据资产详情
// @Security ApiKeyAuth
// @Produce application/json
// @Param id query int true "资产ID"
// @Success 200 {object} response.Response{data=model.LlmopsDataAsset} "获取成功"
// @Router /llmops/asset [get]
func (a *LlmopsApi) GetAsset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	asset, err := service.LlmopsService.GetAsset(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(asset, "获取成功", c)
}

// DeleteAsset 删除数据资产（软删除）
// @Tags Llmops
// @Summary 删除数据资产（软删除）
// @Security ApiKeyAuth
// @Param id query int true "资产ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/asset [delete]
func (a *LlmopsApi) DeleteAsset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteAsset(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ImportDoc 文档采集上传
// @Tags Llmops
// @Summary 文档采集（文本类直读正文 parsed，二进制仅编目 draft）
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce application/json
// @Param file formData file true "文档文件"
// @Param name formData string false "资产名（默认取文件名）"
// @Param tags formData string false "标签(逗号分隔)"
// @Success 200 {object} response.Response{data=model.LlmopsDataAsset} "采集成功"
// @Router /llmops/asset/importDoc [post]
func (a *LlmopsApi) ImportDoc(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("缺少 file 文件字段", c)
		return
	}
	asset, err := service.LlmopsService.ImportDoc(fh, c.PostForm("name"), c.PostForm("tags"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(asset, "采集成功", c)
}

// ImportWeb 网页采集
// @Tags Llmops
// @Summary 网页采集（URL 拉取→HTML 正文抽取→落库 parsed）
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body webImportReq true "采集请求"
// @Success 200 {object} response.Response{data=model.LlmopsDataAsset} "采集成功"
// @Router /llmops/asset/importWeb [post]
func (a *LlmopsApi) ImportWeb(c *gin.Context) {
	var req webImportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	asset, err := service.LlmopsService.ImportWebPage(req.URL, req.Name, req.Tags)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(asset, "采集成功", c)
}

// ImportQa 问答对采集
// @Tags Llmops
// @Summary 问答对采集（JSON/CSV 自动探测，同题去重）
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce application/json
// @Param file formData file true "问答对文件(.json 数组或 .csv 表头 question/answer)"
// @Param name formData string false "资产名"
// @Param tags formData string false "标签(逗号分隔)"
// @Success 200 {object} response.Response{data=map[string]any} "采集成功"
// @Router /llmops/asset/importQa [post]
func (a *LlmopsApi) ImportQa(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("缺少 file 文件字段", c)
		return
	}
	if fh.Size > ImportQaMaxBytes {
		response.FailWithMessage("文件超过 10MB 上限", c)
		return
	}
	f, err := fh.Open()
	if err != nil {
		response.FailWithMessage("读取上传文件失败", c)
		return
	}
	defer f.Close()
	content, err := io.ReadAll(io.LimitReader(f, ImportQaMaxBytes+1))
	if err != nil {
		response.FailWithMessage("读取上传文件失败", c)
		return
	}
	asset, imported, skipped, dup, err := service.LlmopsService.ImportQaPairs(c.PostForm("name"), content, c.PostForm("tags"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"asset": asset, "imported": imported, "skipped": skipped, "dup": dup}, "采集成功", c)
}

// GetDataSources 数据源列表
// @Tags Llmops
// @Summary 结构化数据源列表（密码不回显）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.LlmopsDataSource} "获取成功"
// @Router /llmops/datasource/list [get]
func (a *LlmopsApi) GetDataSources(c *gin.Context) {
	list, err := service.LlmopsService.GetDataSources()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveDataSource 保存数据源
// @Tags Llmops
// @Summary 新建/更新数据源（密码留空不改）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.DataSourceIn true "数据源"
// @Success 200 {object} response.Response{data=model.LlmopsDataSource} "保存成功"
// @Router /llmops/datasource [post]
func (a *LlmopsApi) SaveDataSource(c *gin.Context) {
	var in service.DataSourceIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	ds, err := service.LlmopsService.SaveDataSource(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(ds, "保存成功", c)
}

// DeleteDataSource 删除数据源
// @Tags Llmops
// @Summary 删除数据源（软删除）
// @Security ApiKeyAuth
// @Param id query int true "数据源ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/datasource [delete]
func (a *LlmopsApi) DeleteDataSource(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteDataSource(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// TestDataSource 数据源连接测试
// @Tags Llmops
// @Summary 连接测试（传 id 测已保存；传完整表单直连测试）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.DataSourceIn true "数据源（可仅 id 或完整表单）"
// @Success 200 {object} response.Response{msg=string} "连接成功"
// @Router /llmops/datasource/test [post]
func (a *LlmopsApi) TestDataSource(c *gin.Context) {
	var in service.DataSourceIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var err error
	if in.ID > 0 {
		err = service.LlmopsService.TestDataSource(in.ID)
	} else {
		err = service.LlmopsService.TestDataSourceDirect(&in)
	}
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("连接成功", c)
}

// ListDataSourceTables 库表清单
// @Tags Llmops
// @Summary 库表清单
// @Security ApiKeyAuth
// @Param id query int true "数据源ID"
// @Success 200 {object} response.Response{data=[]string} "获取成功"
// @Router /llmops/datasource/tables [get]
func (a *LlmopsApi) ListDataSourceTables(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	tables, err := service.LlmopsService.ListDataSourceTables(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(tables, "获取成功", c)
}

// PreviewDataSourceTable 采样预览
// @Tags Llmops
// @Summary 表采样预览（默认 20 行，上限 100）
// @Security ApiKeyAuth
// @Param id query int true "数据源ID"
// @Param table query string true "表名"
// @Param limit query int false "行数上限"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /llmops/datasource/preview [get]
func (a *LlmopsApi) PreviewDataSourceTable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	columns, rows, err := service.LlmopsService.PreviewDataSourceTable(uint(id), c.Query("table"), limit)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"columns": columns, "rows": rows}, "获取成功", c)
}

// ImportSemi 半结构化数据编目
// @Tags Llmops
// @Summary 半结构化编目（JSON/YAML/XML 自动探测，顶层字段抽取入 metaJSON）
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce application/json
// @Param file formData file true "半结构化文件(.json/.yaml/.xml)"
// @Param name formData string false "资产名"
// @Param tags formData string false "标签(逗号分隔)"
// @Success 200 {object} response.Response{data=model.LlmopsDataAsset} "编目成功"
// @Router /llmops/asset/importSemi [post]
func (a *LlmopsApi) ImportSemi(c *gin.Context) {
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
	content, err := io.ReadAll(io.LimitReader(f, 10<<20+1))
	if err != nil {
		response.FailWithMessage("读取上传文件失败", c)
		return
	}
	asset, err := service.LlmopsService.ImportSemiStruct(c.PostForm("name"), content, c.PostForm("tags"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(asset, "编目成功", c)
}

// SaveCategory 保存编目分类
// @Tags Llmops
// @Summary 新建/更新编目分类（父分类存在性+环检测）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.LlmopsCategory true "分类"
// @Success 200 {object} response.Response{msg=string} "保存成功"
// @Router /llmops/category [post]
func (a *LlmopsApi) SaveCategory(c *gin.Context) {
	var cat model.LlmopsCategory
	if err := c.ShouldBindJSON(&cat); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.LlmopsService.SaveCategory(&cat); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(cat, "保存成功", c)
}

// GetCategoryTree 分类树
// @Tags Llmops
// @Summary 编目分类树（孤儿/环引用降级挂根）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]service.CategoryNode} "获取成功"
// @Router /llmops/category/tree [get]
func (a *LlmopsApi) GetCategoryTree(c *gin.Context) {
	tree, err := service.LlmopsService.GetCategoryTree()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(tree, "获取成功", c)
}

// DeleteCategory 删除分类
// @Tags Llmops
// @Summary 删除编目分类（有子分类/有资产引用拒绝）
// @Security ApiKeyAuth
// @Param id query int true "分类ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/category [delete]
func (a *LlmopsApi) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteCategory(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// CategorizeAsset 资产挂分类/打标签
// @Tags Llmops
// @Summary 资产挂分类/打标签（记事件流水）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body categorizeReq true "挂载请求"
// @Success 200 {object} response.Response{msg=string} "操作成功"
// @Router /llmops/asset/categorize [post]
func (a *LlmopsApi) CategorizeAsset(c *gin.Context) {
	var req categorizeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.LlmopsService.AssignAssetCategory(req.AssetId, req.CategoryId, req.Tags); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}

// GetAssetEvents 资产事件流水
// @Tags Llmops
// @Summary 资产生命周期事件流水（倒序分页）
// @Security ApiKeyAuth
// @Param assetId query int true "资产ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /llmops/asset/events [get]
func (a *LlmopsApi) GetAssetEvents(c *gin.Context) {
	assetId, err := strconv.ParseUint(c.Query("assetId"), 10, 64)
	if err != nil || assetId == 0 {
		response.FailWithMessage("assetId 不能为空", c)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.LlmopsService.GetAssetEvents(uint(assetId), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}
