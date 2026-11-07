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

// transcribeReq 媒体转写请求
type transcribeReq struct {
	AssetId uint `json:"assetId" binding:"required"`
}

// idReq 通用 ID 请求体
type idReq struct {
	ID uint `json:"id" binding:"required"`
}

// ingestIn 已移至 service.IngestIn（知识入库请求）

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
// @Summary 数据资产列表（类型/状态/关键字/标签/分类/时间范围多路组合检索）
// @Security ApiKeyAuth
// @Produce application/json
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param dataType query string false "数据类型(doc/image/audio/video/webpage/qa/structured/semistructured)"
// @Param status query string false "状态(draft/parsed/cataloged/offline)"
// @Param keyword query string false "名称/标签关键字"
// @Param tags query string false "标签过滤"
// @Param categoryId query int false "分类ID"
// @Param begin query string false "创建时间起(RFC3339 或 2006-01-02)"
// @Param end query string false "创建时间止"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /llmops/asset/list [get]
func (a *LlmopsApi) GetAssets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	categoryId, _ := strconv.ParseUint(c.Query("categoryId"), 10, 64)
	list, total, err := service.LlmopsService.GetAssetList(page, pageSize,
		c.Query("dataType"), c.Query("status"), c.Query("keyword"), c.Query("tags"),
		uint(categoryId), c.Query("begin"), c.Query("end"))
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

// ImportMedia 媒体编目
// @Tags Llmops
// @Summary 音视频/图片元数据编目（转写待 ASR 适配器）
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce application/json
// @Param file formData file true "媒体文件"
// @Param name formData string false "资产名"
// @Param tags formData string false "标签(逗号分隔)"
// @Success 200 {object} response.Response{data=model.LlmopsDataAsset} "编目成功"
// @Router /llmops/asset/importMedia [post]
func (a *LlmopsApi) ImportMedia(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("缺少 file 文件字段", c)
		return
	}
	asset, err := service.LlmopsService.ImportMedia(fh, c.PostForm("name"), c.PostForm("tags"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(asset, "编目成功", c)
}

// TranscribeMedia 媒体转写
// @Tags Llmops
// @Summary 音频/视频 ASR 转写（适配器未配置则明确报错）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body transcribeReq true "转写请求"
// @Success 200 {object} response.Response{data=model.LlmopsDataAsset} "转写成功"
// @Router /llmops/asset/transcribe [post]
func (a *LlmopsApi) TranscribeMedia(c *gin.Context) {
	var req transcribeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	asset, err := service.LlmopsService.TranscribeMedia(req.AssetId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(asset, "转写成功", c)
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

// GetModels 模型注册列表
// @Tags Llmops
// @Summary 模型注册列表（密钥不回显）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.LlmopsModel} "获取成功"
// @Router /llmops/model/list [get]
func (a *LlmopsApi) GetModels(c *gin.Context) {
	list, err := service.LlmopsService.GetModels()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveModel 保存模型注册
// @Tags Llmops
// @Summary 新建/更新模型注册（API Key 留空不改；DefaultFor 同能力互斥）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.ModelIn true "模型"
// @Success 200 {object} response.Response{data=model.LlmopsModel} "保存成功"
// @Router /llmops/model [post]
func (a *LlmopsApi) SaveModel(c *gin.Context) {
	var in service.ModelIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	m, err := service.LlmopsService.SaveModel(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(m, "保存成功", c)
}

// DeleteModel 删除模型注册
// @Tags Llmops
// @Summary 删除模型注册（软删除）
// @Security ApiKeyAuth
// @Param id query int true "模型ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/model [delete]
func (a *LlmopsApi) DeleteModel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteModel(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// TestModel 在线测试
// @Tags Llmops
// @Summary 模型在线测试（chat/embed/rerank 三协议，成败均记调用日志）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.TestProtocolIn true "测试请求"
// @Success 200 {object} response.Response{data=map[string]any} "测试完成"
// @Router /llmops/model/test [post]
func (a *LlmopsApi) TestModel(c *gin.Context) {
	var in service.TestProtocolIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := service.LlmopsService.TestModelProtocol(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(result, "测试完成", c)
}

// GetModelCallLogs 调用日志
// @Tags Llmops
// @Summary 模型调用日志（模型/成败过滤，倒序分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Param modelId query int false "模型ID"
// @Param ok query string false "true/false"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /llmops/model/callLogs [get]
func (a *LlmopsApi) GetModelCallLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	modelId, _ := strconv.ParseUint(c.Query("modelId"), 10, 64)
	list, total, err := service.LlmopsService.GetModelCallLogs(page, pageSize, uint(modelId), c.Query("ok"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetDeploys 部署列表
// @Tags Llmops
// @Summary 模型服务部署列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.LlmopsModelDeploy} "获取成功"
// @Router /llmops/deploy/list [get]
func (a *LlmopsApi) GetDeploys(c *gin.Context) {
	list, err := service.LlmopsService.GetDeploys()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveDeployPlan 保存部署计划
// @Tags Llmops
// @Summary 新建/更新部署计划（镜像必填，端口/环境即检）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.DeployIn true "部署计划"
// @Success 200 {object} response.Response{data=model.LlmopsModelDeploy} "保存成功"
// @Router /llmops/deploy [post]
func (a *LlmopsApi) SaveDeployPlan(c *gin.Context) {
	var in service.DeployIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	d, err := service.LlmopsService.SaveDeployPlan(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(d, "保存成功", c)
}

// ExecuteDeploy 执行部署
// @Tags Llmops
// @Summary 一键部署（建容器+健康探测）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body idReq true "部署ID"
// @Success 200 {object} response.Response{data=model.LlmopsModelDeploy} "部署完成"
// @Router /llmops/deploy/execute [post]
func (a *LlmopsApi) ExecuteDeploy(c *gin.Context) {
	var req idReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	d, err := service.LlmopsService.ExecuteDeploy(req.ID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(d, "部署完成", c)
}

// DestroyDeploy 销毁部署
// @Tags Llmops
// @Summary 销毁部署（强删容器回置 stopped）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body idReq true "部署ID"
// @Success 200 {object} response.Response{data=model.LlmopsModelDeploy} "销毁完成"
// @Router /llmops/deploy/destroy [post]
func (a *LlmopsApi) DestroyDeploy(c *gin.Context) {
	var req idReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	d, err := service.LlmopsService.DestroyDeploy(req.ID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(d, "销毁完成", c)
}

// DeleteDeploy 删除部署计划
// @Tags Llmops
// @Summary 删除部署计划（运行中拒绝）
// @Security ApiKeyAuth
// @Param id query int true "部署ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/deploy [delete]
func (a *LlmopsApi) DeleteDeploy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteDeploy(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// SelectModel 能力调度
// @Tags Llmops
// @Summary 按能力调度模型（DefaultFor 优先，其余按近 10 次成功调用时延均值）
// @Security ApiKeyAuth
// @Produce application/json
// @Param capability query string true "能力标签(chat/embed/rerank)"
// @Success 200 {object} response.Response{data=model.LlmopsModel} "获取成功"
// @Router /llmops/model/select [get]
func (a *LlmopsApi) SelectModel(c *gin.Context) {
	m, err := service.LlmopsService.SelectModel(c.Query("capability"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(m, "获取成功", c)
}

// GetKnowledges 知识库列表
// @Tags Llmops
// @Summary 知识库列表（vector/fulltext/graph/qa 四类）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]model.LlmopsKnowledge} "获取成功"
// @Router /llmops/kb/list [get]
func (a *LlmopsApi) GetKnowledges(c *gin.Context) {
	list, err := service.LlmopsService.GetKnowledges()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveKnowledge 保存知识库
// @Tags Llmops
// @Summary 新建/更新知识库
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.KnowledgeIn true "知识库"
// @Success 200 {object} response.Response{data=model.LlmopsKnowledge} "保存成功"
// @Router /llmops/kb [post]
func (a *LlmopsApi) SaveKnowledge(c *gin.Context) {
	var in service.KnowledgeIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	kb, err := service.LlmopsService.SaveKnowledge(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(kb, "保存成功", c)
}

// DeleteKnowledge 删除知识库
// @Tags Llmops
// @Summary 删除知识库（级联软删文档）
// @Security ApiKeyAuth
// @Param id query int true "知识库ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/kb [delete]
func (a *LlmopsApi) DeleteKnowledge(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteKnowledge(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// IngestKnowledgeDoc 文档入库
// @Tags Llmops
// @Summary 文档入库（资产 rawText 或内联文本，原文留档）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.IngestIn true "入库请求"
// @Success 200 {object} response.Response{data=model.LlmopsKnowledgeDoc} "入库成功"
// @Router /llmops/kb/ingest [post]
func (a *LlmopsApi) IngestKnowledgeDoc(c *gin.Context) {
	var in service.IngestIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	doc, err := service.LlmopsService.IngestDocument(in.KbId, in.AssetId, in.Title, in.Text)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(doc, "入库成功", c)
}

// GetKnowledgeDocs 库内文档列表
// @Tags Llmops
// @Summary 库内文档列表（不含正文，分页）
// @Security ApiKeyAuth
// @Param kbId query int true "知识库ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /llmops/kb/docs [get]
func (a *LlmopsApi) GetKnowledgeDocs(c *gin.Context) {
	kbId, err := strconv.ParseUint(c.Query("kbId"), 10, 64)
	if err != nil || kbId == 0 {
		response.FailWithMessage("kbId 不能为空", c)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.LlmopsService.GetKnowledgeDocs(uint(kbId), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// DeleteKnowledgeDoc 删除库内文档
// @Tags Llmops
// @Summary 删除库内文档（软删除）
// @Security ApiKeyAuth
// @Param id query int true "文档ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/kb/doc [delete]
func (a *LlmopsApi) DeleteKnowledgeDoc(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteKnowledgeDoc(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// IndexVectorDoc 向量索引
// @Tags Llmops
// @Summary 文档向量索引（切块→embed 模型/哈希 fallback→落库）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.VectorIndexIn true "索引请求"
// @Success 200 {object} response.Response{data=service.IndexResult} "索引完成"
// @Router /llmops/kb/vector/index [post]
func (a *LlmopsApi) IndexVectorDoc(c *gin.Context) {
	var in service.VectorIndexIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := service.LlmopsService.IndexDocToVector(in.KbId, in.DocId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(result, "索引完成", c)
}

// SearchVector 向量检索
// @Tags Llmops
// @Summary 向量语义检索（同 Source 余弦 top-k）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.VectorSearchIn true "检索请求"
// @Success 200 {object} response.Response{data=map[string]any} "检索完成"
// @Router /llmops/kb/vector/search [post]
func (a *LlmopsApi) SearchVector(c *gin.Context) {
	var in service.VectorSearchIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	hits, source, err := service.LlmopsService.SearchVector(in.KbId, in.Query, in.TopK)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"hits": hits, "source": source}, "检索完成", c)
}

// IndexFulltextDoc 全文索引
// @Tags Llmops
// @Summary 文档全文索引（切块→bigram 词元落库）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.VectorIndexIn true "索引请求"
// @Success 200 {object} response.Response{data=service.IndexResult} "索引完成"
// @Router /llmops/kb/fulltext/index [post]
func (a *LlmopsApi) IndexFulltextDoc(c *gin.Context) {
	var in service.VectorIndexIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := service.LlmopsService.IndexDocToFulltext(in.KbId, in.DocId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(result, "索引完成", c)
}

// SearchFulltext 全文检索
// @Tags Llmops
// @Summary 全文检索（bigram + 简化 BM25 top-k）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.VectorSearchIn true "检索请求"
// @Success 200 {object} response.Response{data=map[string]any} "检索完成"
// @Router /llmops/kb/fulltext/search [post]
func (a *LlmopsApi) SearchFulltext(c *gin.Context) {
	var in service.VectorSearchIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	hits, err := service.LlmopsService.SearchFulltext(in.KbId, in.Query, in.TopK)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"hits": hits}, "检索完成", c)
}

// SearchQa 问答检索
// @Tags Llmops
// @Summary 问答对检索（精确归一命中优先→编辑相似度阈值）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.QaSearchIn true "检索请求"
// @Success 200 {object} response.Response{data=map[string]any} "检索完成"
// @Router /llmops/kb/qa/search [post]
func (a *LlmopsApi) SearchQa(c *gin.Context) {
	var in service.QaSearchIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	hits, err := service.LlmopsService.SearchQa(in.KbId, in.Query, in.TopK, in.MinScore)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"hits": hits}, "检索完成", c)
}

// SaveGraphNode 保存图谱节点
// @Tags Llmops
// @Summary 新建/更新图谱节点（库内同名去重）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.LlmopsGraphNode true "节点"
// @Success 200 {object} response.Response{msg=string} "保存成功"
// @Router /llmops/kb/graph/node [post]
func (a *LlmopsApi) SaveGraphNode(c *gin.Context) {
	var n model.LlmopsGraphNode
	if err := c.ShouldBindJSON(&n); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.LlmopsService.SaveGraphNode(&n); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(n, "保存成功", c)
}

// GetGraphNodes 图谱节点列表
// @Tags Llmops
// @Summary 库内图谱节点列表
// @Security ApiKeyAuth
// @Param kbId query int true "知识库ID"
// @Success 200 {object} response.Response{data=[]model.LlmopsGraphNode} "获取成功"
// @Router /llmops/kb/graph/nodes [get]
func (a *LlmopsApi) GetGraphNodes(c *gin.Context) {
	kbId, err := strconv.ParseUint(c.Query("kbId"), 10, 64)
	if err != nil || kbId == 0 {
		response.FailWithMessage("kbId 不能为空", c)
		return
	}
	list, err := service.LlmopsService.GetGraphNodes(uint(kbId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// DeleteGraphNode 删除图谱节点
// @Tags Llmops
// @Summary 删除节点（级联删关联边）
// @Security ApiKeyAuth
// @Param id query int true "节点ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/kb/graph/node [delete]
func (a *LlmopsApi) DeleteGraphNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteGraphNode(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// SaveGraphEdge 保存图谱边
// @Tags Llmops
// @Summary 新建图谱边（起终点同库校验）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.LlmopsGraphEdge true "边"
// @Success 200 {object} response.Response{msg=string} "保存成功"
// @Router /llmops/kb/graph/edge [post]
func (a *LlmopsApi) SaveGraphEdge(c *gin.Context) {
	var e model.LlmopsGraphEdge
	if err := c.ShouldBindJSON(&e); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.LlmopsService.SaveGraphEdge(&e); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(e, "保存成功", c)
}

// GetGraphEdges 图谱边列表
// @Tags Llmops
// @Summary 库内图谱边列表
// @Security ApiKeyAuth
// @Param kbId query int true "知识库ID"
// @Success 200 {object} response.Response{data=[]model.LlmopsGraphEdge} "获取成功"
// @Router /llmops/kb/graph/edges [get]
func (a *LlmopsApi) GetGraphEdges(c *gin.Context) {
	kbId, err := strconv.ParseUint(c.Query("kbId"), 10, 64)
	if err != nil || kbId == 0 {
		response.FailWithMessage("kbId 不能为空", c)
		return
	}
	list, err := service.LlmopsService.GetGraphEdges(uint(kbId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// DeleteGraphEdge 删除图谱边
// @Tags Llmops
// @Summary 删除图谱边
// @Security ApiKeyAuth
// @Param id query int true "边ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/kb/graph/edge [delete]
func (a *LlmopsApi) DeleteGraphEdge(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteGraphEdge(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GraphMultiHopQuery 多跳推理
// @Tags Llmops
// @Summary 图谱多跳推理（BFS 路径枚举，hops 1-5，路径上限 100）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.GraphQueryIn true "推理请求"
// @Success 200 {object} response.Response{data=map[string]any} "推理完成"
// @Router /llmops/kb/graph/query [post]
func (a *LlmopsApi) GraphMultiHopQuery(c *gin.Context) {
	var in service.GraphQueryIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	paths, err := service.LlmopsService.GraphMultiHopQuery(in.KbId, in.StartNodeId, in.MaxHops, in.Relation)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"paths": paths}, "推理完成", c)
}

// GenerateDocSummary 生成文档摘要
// @Tags Llmops
// @Summary 创建摘要生成任务（异步，chat 模型；同文档覆盖旧任务）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.SummaryGenIn true "生成请求"
// @Success 200 {object} response.Response{data=model.LlmopsDocSummary} "任务已创建"
// @Router /llmops/kb/summary/generate [post]
func (a *LlmopsApi) GenerateDocSummary(c *gin.Context) {
	var in service.SummaryGenIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	task, err := service.LlmopsService.CreateSummaryTask(in.KbId, in.DocId, in.ModelId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(task, "任务已创建", c)
}

// GetDocSummary 查询文档摘要
// @Tags Llmops
// @Summary 查询摘要任务状态与结果
// @Security ApiKeyAuth
// @Param kbId query int true "知识库ID"
// @Param docId query int true "文档ID"
// @Success 200 {object} response.Response{data=model.LlmopsDocSummary} "获取成功"
// @Router /llmops/kb/summary [get]
func (a *LlmopsApi) GetDocSummary(c *gin.Context) {
	kbId, err1 := strconv.ParseUint(c.Query("kbId"), 10, 64)
	docId, err2 := strconv.ParseUint(c.Query("docId"), 10, 64)
	if err1 != nil || err2 != nil || kbId == 0 || docId == 0 {
		response.FailWithMessage("kbId/docId 不能为空", c)
		return
	}
	task, err := service.LlmopsService.GetDocSummary(uint(kbId), uint(docId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(task, "获取成功", c)
}

// SearchSummaries 摘要检索
// @Tags Llmops
// @Summary 摘要检索（bigram+BM25，仅已完成摘要）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.VectorSearchIn true "检索请求"
// @Success 200 {object} response.Response{data=map[string]any} "检索完成"
// @Router /llmops/kb/summary/search [post]
func (a *LlmopsApi) SearchSummaries(c *gin.Context) {
	var in service.VectorSearchIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	hits, err := service.LlmopsService.SearchSummaries(in.KbId, in.Query, in.TopK)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"hits": hits}, "检索完成", c)
}

// UnifiedSearch 统一检索入口
// @Tags Llmops
// @Summary 统一检索（多库 hybrid/单路由模式，成败均留痕）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.UnifiedSearchIn true "检索请求"
// @Success 200 {object} response.Response{data=map[string]any} "检索完成"
// @Router /llmops/knowledge/search [post]
func (a *LlmopsApi) UnifiedSearch(c *gin.Context) {
	var in service.UnifiedSearchIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	hits, report, err := service.LlmopsService.UnifiedSearch(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"hits": hits, "report": report}, "检索完成", c)
}

// GetSearchLogs 检索日志
// @Tags Llmops
// @Summary 统一检索调用日志（倒序分页）
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /llmops/knowledge/searchLogs [get]
func (a *LlmopsApi) GetSearchLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.LlmopsService.GetSearchLogs(page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetTrainJobs 训练任务列表
// @Tags Llmops
// @Summary 训练/调优任务列表（类型/状态过滤，分页）
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Param type query string false "sft_full/sft_lora/reward/rl/eval/compress"
// @Param status query string false "状态"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /llmops/train/list [get]
func (a *LlmopsApi) GetTrainJobs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	list, total, err := service.LlmopsService.GetTrainJobs(page, pageSize, c.Query("type"), c.Query("status"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// SaveTrainJob 保存训练任务
// @Tags Llmops
// @Summary 新建/更新训练任务（草稿态可编辑，超参 JSON 即检）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.TrainIn true "任务"
// @Success 200 {object} response.Response{data=model.LlmopsTrainJob} "保存成功"
// @Router /llmops/train [post]
func (a *LlmopsApi) SaveTrainJob(c *gin.Context) {
	var in service.TrainIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	job, err := service.LlmopsService.SaveTrainJob(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(job, "保存成功", c)
}

// TransitionTrainJob 任务状态流转
// @Tags Llmops
// @Summary 训练任务状态流转（状态机校验）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.TrainTransitionIn true "流转请求"
// @Success 200 {object} response.Response{data=model.LlmopsTrainJob} "流转成功"
// @Router /llmops/train/transition [post]
func (a *LlmopsApi) TransitionTrainJob(c *gin.Context) {
	var in service.TrainTransitionIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	job, err := service.LlmopsService.TransitionTrainJob(in.ID, in.To, in.Reason)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(job, "流转成功", c)
}

// DeleteTrainJob 删除训练任务
// @Tags Llmops
// @Summary 删除训练任务（运行/排队中拒绝）
// @Security ApiKeyAuth
// @Param id query int true "任务ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/train [delete]
func (a *LlmopsApi) DeleteTrainJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteTrainJob(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DryRunTrain 任务 dry-run
// @Tags Llmops
// @Summary 训练任务 dry-run（渲染命令落日志，不执行）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body idReq true "任务ID"
// @Success 200 {object} response.Response{data=model.LlmopsTrainJob} "渲染完成"
// @Router /llmops/train/dryRun [post]
func (a *LlmopsApi) DryRunTrain(c *gin.Context) {
	var req idReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	job, err := service.LlmopsService.RunTrainDryRun(req.ID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(job, "渲染完成", c)
}

// StartTrainJob 启动任务
// @Tags Llmops
// @Summary 启动训练任务（dry-run 执行器：渲染+置 running，真实后端待接入）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body idReq true "任务ID"
// @Success 200 {object} response.Response{data=model.LlmopsTrainJob} "已启动"
// @Router /llmops/train/start [post]
func (a *LlmopsApi) StartTrainJob(c *gin.Context) {
	var req idReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	job, err := service.LlmopsService.StartTrainJob(req.ID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(job, "已启动", c)
}

// StopTrainJob 停止任务
// @Tags Llmops
// @Summary 停止训练任务（running→stopped）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.TrainStopIn true "停止请求"
// @Success 200 {object} response.Response{data=model.LlmopsTrainJob} "已停止"
// @Router /llmops/train/stop [post]
func (a *LlmopsApi) StopTrainJob(c *gin.Context) {
	var in service.TrainStopIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	job, err := service.LlmopsService.StopTrainJob(in.ID, in.Reason)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(job, "已停止", c)
}

// GetTrainJobLog 任务日志
// @Tags Llmops
// @Summary 训练任务执行日志
// @Security ApiKeyAuth
// @Param id query int true "任务ID"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /llmops/train/log [get]
func (a *LlmopsApi) GetTrainJobLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	logText, status, err := service.LlmopsService.GetTrainJobLog(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"log": logText, "status": status}, "获取成功", c)
}

// GetEvalRuns 评估任务列表
// @Tags Llmops
// @Summary 模型评估任务列表（分页）
// @Security ApiKeyAuth
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /llmops/eval/list [get]
func (a *LlmopsApi) GetEvalRuns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	list, total, err := service.LlmopsService.GetEvalRuns(page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// SaveEvalRun 保存评估任务
// @Tags Llmops
// @Summary 新建/更新评估任务（指标/样本即检）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.LlmopsEvalRun true "评估任务"
// @Success 200 {object} response.Response{data=model.LlmopsEvalRun} "保存成功"
// @Router /llmops/eval [post]
func (a *LlmopsApi) SaveEvalRun(c *gin.Context) {
	var in model.LlmopsEvalRun
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	run, err := service.LlmopsService.SaveEvalRun(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(run, "保存成功", c)
}

// RunEval 执行评估
// @Tags Llmops
// @Summary 执行模型评估（同步小批量 ≤50 样本，明细留档）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body idReq true "评估任务ID"
// @Success 200 {object} response.Response{data=model.LlmopsEvalRun} "评估完成"
// @Router /llmops/eval/run [post]
func (a *LlmopsApi) RunEval(c *gin.Context) {
	var req idReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	run, err := service.LlmopsService.RunEval(req.ID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(run, "评估完成", c)
}

// GetEvalResults 评估明细
// @Tags Llmops
// @Summary 评估明细（按样本序号）
// @Security ApiKeyAuth
// @Param runId query int true "评估任务ID"
// @Success 200 {object} response.Response{data=[]model.LlmopsEvalResult} "获取成功"
// @Router /llmops/eval/results [get]
func (a *LlmopsApi) GetEvalResults(c *gin.Context) {
	runId, err := strconv.ParseUint(c.Query("runId"), 10, 64)
	if err != nil || runId == 0 {
		response.FailWithMessage("runId 不能为空", c)
		return
	}
	list, err := service.LlmopsService.GetEvalResults(uint(runId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// GetArtifacts 产物列表
// @Tags Llmops
// @Summary 训练/压缩产物列表
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]model.LlmopsArtifact} "获取成功"
// @Router /llmops/artifact/list [get]
func (a *LlmopsApi) GetArtifacts(c *gin.Context) {
	list, err := service.LlmopsService.GetArtifacts()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// RegisterArtifact 产物登记
// @Tags Llmops
// @Summary 登记训练/压缩产物（D07 预测服务注册对接面）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.LlmopsArtifact true "产物"
// @Success 200 {object} response.Response{data=model.LlmopsArtifact} "登记成功"
// @Router /llmops/artifact [post]
func (a *LlmopsApi) RegisterArtifact(c *gin.Context) {
	var in model.LlmopsArtifact
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	art, err := service.LlmopsService.RegisterArtifact(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(art, "登记成功", c)
}

// DeleteArtifact 删除产物
// @Tags Llmops
// @Summary 删除产物
// @Security ApiKeyAuth
// @Param id query int true "产物ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/artifact [delete]
func (a *LlmopsApi) DeleteArtifact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteArtifact(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// RegisterArtifactModel 产物注册为模型服务
// @Tags Llmops
// @Summary 训练/压缩产物一键注册为模型服务（走模型注册表全量校验）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.ArtifactModelIn true "注册请求"
// @Success 200 {object} response.Response{data=model.LlmopsModel} "注册成功"
// @Router /llmops/artifact/registerModel [post]
func (a *LlmopsApi) RegisterArtifactModel(c *gin.Context) {
	var in service.ArtifactModelIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	m, err := service.LlmopsService.RegisterArtifactAsModel(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(m, "注册成功", c)
}

// GetAgents 智能体列表
// @Tags Llmops
// @Summary 智能体列表（五种构建模式）
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]model.LlmopsAgent} "获取成功"
// @Router /llmops/agent/list [get]
func (a *LlmopsApi) GetAgents(c *gin.Context) {
	list, err := service.LlmopsService.GetAgents()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveAgent 保存智能体
// @Tags Llmops
// @Summary 新建/更新智能体（模式白名单/配置 JSON 对象即检）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.AgentIn true "智能体"
// @Success 200 {object} response.Response{data=model.LlmopsAgent} "保存成功"
// @Router /llmops/agent [post]
func (a *LlmopsApi) SaveAgent(c *gin.Context) {
	var in service.AgentIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	ag, err := service.LlmopsService.SaveAgent(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(ag, "保存成功", c)
}

// SetAgentStage 生命周期阶段流转
// @Tags Llmops
// @Summary 智能体阶段流转（六阶段白名单）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.AgentStageIn true "阶段请求"
// @Success 200 {object} response.Response{data=model.LlmopsAgent} "流转成功"
// @Router /llmops/agent/stage [post]
func (a *LlmopsApi) SetAgentStage(c *gin.Context) {
	var in service.AgentStageIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	ag, err := service.LlmopsService.SetAgentStage(in.ID, in.Stage)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(ag, "流转成功", c)
}

// PublishAgent 发布智能体
// @Tags Llmops
// @Summary 发布智能体（配置快照落版本表，版本号自增）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.AgentPublishIn true "发布请求"
// @Success 200 {object} response.Response{data=model.LlmopsAgentVersion} "发布成功"
// @Router /llmops/agent/publish [post]
func (a *LlmopsApi) PublishAgent(c *gin.Context) {
	var in service.AgentPublishIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	ver, err := service.LlmopsService.PublishAgent(in.ID, in.Remark)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(ver, "发布成功", c)
}

// GetAgentVersions 版本列表
// @Tags Llmops
// @Summary 智能体版本快照列表（倒序）
// @Security ApiKeyAuth
// @Param agentId query int true "智能体ID"
// @Success 200 {object} response.Response{data=[]model.LlmopsAgentVersion} "获取成功"
// @Router /llmops/agent/versions [get]
func (a *LlmopsApi) GetAgentVersions(c *gin.Context) {
	agentId, err := strconv.ParseUint(c.Query("agentId"), 10, 64)
	if err != nil || agentId == 0 {
		response.FailWithMessage("agentId 不能为空", c)
		return
	}
	list, err := service.LlmopsService.GetAgentVersions(uint(agentId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// DeleteAgent 删除智能体
// @Tags Llmops
// @Summary 删除智能体（软删除）
// @Security ApiKeyAuth
// @Param id query int true "智能体ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/agent [delete]
func (a *LlmopsApi) DeleteAgent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteAgent(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// CreateAgentSession 创建会话
// @Tags Llmops
// @Summary 创建智能体会话
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.SessionIn true "会话请求"
// @Success 200 {object} response.Response{data=model.LlmopsAgentSession} "创建成功"
// @Router /llmops/agent/session [post]
func (a *LlmopsApi) CreateAgentSession(c *gin.Context) {
	var in service.SessionIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	sess, err := service.LlmopsService.CreateSession(in.AgentId, in.UserKey, in.Title)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(sess, "创建成功", c)
}

// GetAgentMessages 会话消息
// @Tags Llmops
// @Summary 会话消息（正序，近 N 条）
// @Security ApiKeyAuth
// @Param sessionId query int true "会话ID"
// @Success 200 {object} response.Response{data=[]model.LlmopsAgentMessage} "获取成功"
// @Router /llmops/agent/messages [get]
func (a *LlmopsApi) GetAgentMessages(c *gin.Context) {
	sessionId, err := strconv.ParseUint(c.Query("sessionId"), 10, 64)
	if err != nil || sessionId == 0 {
		response.FailWithMessage("sessionId 不能为空", c)
		return
	}
	list, err := service.LlmopsService.GetSessionMessages(uint(sessionId), 50)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// AgentChat 会话对话
// @Tags Llmops
// @Summary 智能体单轮对话（qa 直答；其余先意图拆解后作答）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.AgentChatIn true "对话请求"
// @Success 200 {object} response.Response{data=map[string]any} "对话完成"
// @Router /llmops/agent/chat [post]
func (a *LlmopsApi) AgentChat(c *gin.Context) {
	var in service.AgentChatIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	reply, plan, err := service.LlmopsService.AgentChat(in.SessionId, in.Message)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"reply": reply, "plan": plan}, "对话完成", c)
}

// SaveAgentMemory 保存记忆
// @Tags Llmops
// @Summary 保存智能体长期记忆
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.MemoryIn true "记忆"
// @Success 200 {object} response.Response{data=model.LlmopsAgentMemory} "保存成功"
// @Router /llmops/agent/memory [post]
func (a *LlmopsApi) SaveAgentMemory(c *gin.Context) {
	var in service.MemoryIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	m, err := service.LlmopsService.SaveMemory(in.AgentId, in.UserKey, in.Content, in.Source)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(m, "保存成功", c)
}

// ListAgentMemories 记忆列表
// @Tags Llmops
// @Summary 长期记忆列表（用户专属+全局）
// @Security ApiKeyAuth
// @Param agentId query int true "智能体ID"
// @Param userKey query string false "用户标识"
// @Success 200 {object} response.Response{data=[]model.LlmopsAgentMemory} "获取成功"
// @Router /llmops/agent/memory/list [get]
func (a *LlmopsApi) ListAgentMemories(c *gin.Context) {
	agentId, err := strconv.ParseUint(c.Query("agentId"), 10, 64)
	if err != nil || agentId == 0 {
		response.FailWithMessage("agentId 不能为空", c)
		return
	}
	list, err := service.LlmopsService.ListMemories(uint(agentId), c.Query("userKey"), 50)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// DeleteAgentMemory 删除记忆
// @Tags Llmops
// @Summary 删除长期记忆
// @Security ApiKeyAuth
// @Param id query int true "记忆ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/agent/memory [delete]
func (a *LlmopsApi) DeleteAgentMemory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteMemory(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// InjectAgentMemories 记忆检索注入
// @Tags Llmops
// @Summary 按查询检索相关记忆（BM25 top-k 上下文块）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.MemoryInjectIn true "注入请求"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /llmops/agent/memory/inject [post]
func (a *LlmopsApi) InjectAgentMemories(c *gin.Context) {
	var in service.MemoryInjectIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	ctxText, err := service.LlmopsService.InjectMemories(in.AgentId, in.UserKey, in.Query, in.TopK)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"context": ctxText}, "获取成功", c)
}

// ExtractAgentMemory 自动记忆抽取
// @Tags Llmops
// @Summary LLM 抽取对话中的可记忆事实（无则不写）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.MemoryExtractIn true "抽取请求"
// @Success 200 {object} response.Response{data=map[string]any} "抽取完成"
// @Router /llmops/agent/memory/extract [post]
func (a *LlmopsApi) ExtractAgentMemory(c *gin.Context) {
	var in service.MemoryExtractIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	fact, err := service.LlmopsService.ExtractAutoMemory(in.AgentId, in.UserKey, in.UserMsg, in.Reply, in.ModelId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"fact": fact}, "抽取完成", c)
}

// AgentMultiRun 多智能体运行
// @Tags Llmops
// @Summary 多智能体编排运行（sequential 串联/converge 并行汇聚）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.MultiRunIn true "运行请求"
// @Success 200 {object} response.Response{data=service.MultiRunResult} "运行完成"
// @Router /llmops/agent/multiRun [post]
func (a *LlmopsApi) AgentMultiRun(c *gin.Context) {
	var in service.MultiRunIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	res, err := service.LlmopsService.AgentMultiRun(in.AgentId, in.Input)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(res, "运行完成", c)
}

// AgentDebugChat 调试对话
// @Tags Llmops
// @Summary 智能体调试对话（非 qa 走工具回路，trace 留档）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.AgentChatIn true "对话请求"
// @Success 200 {object} response.Response{data=map[string]any} "对话完成"
// @Router /llmops/agent/debugChat [post]
func (a *LlmopsApi) AgentDebugChat(c *gin.Context) {
	var in service.AgentChatIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	reply, plan, traces, err := service.LlmopsService.AgentDebugChat(in.SessionId, in.Message)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"reply": reply, "plan": plan, "traces": traces}, "对话完成", c)
}

// ListAgentSessions 会话列表
// @Tags Llmops
// @Summary 智能体会话列表（调试历史）
// @Security ApiKeyAuth
// @Param agentId query int true "智能体ID"
// @Param userKey query string false "用户标识"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /llmops/agent/sessions [get]
func (a *LlmopsApi) ListAgentSessions(c *gin.Context) {
	agentId, err := strconv.ParseUint(c.Query("agentId"), 10, 64)
	if err != nil || agentId == 0 {
		response.FailWithMessage("agentId 不能为空", c)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.LlmopsService.ListAgentSessions(uint(agentId), c.Query("userKey"), page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetAgentEvals 智能体评测列表
// @Tags Llmops
// @Summary 智能体评测任务列表
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]model.LlmopsAgentEval} "获取成功"
// @Router /llmops/agent/eval/list [get]
func (a *LlmopsApi) GetAgentEvals(c *gin.Context) {
	list, err := service.LlmopsService.GetAgentEvals()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveAgentEval 保存智能体评测
// @Tags Llmops
// @Summary 新建/更新智能体评测（评分配置即检）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.LlmopsAgentEval true "评测任务"
// @Success 200 {object} response.Response{data=model.LlmopsAgentEval} "保存成功"
// @Router /llmops/agent/eval [post]
func (a *LlmopsApi) SaveAgentEval(c *gin.Context) {
	var in model.LlmopsAgentEval
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	ev, err := service.LlmopsService.SaveAgentEval(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(ev, "保存成功", c)
}

// RunAgentEval 执行智能体评测
// @Tags Llmops
// @Summary 执行智能体评测（≤30 样本，双轨评分）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body idReq true "评测ID"
// @Success 200 {object} response.Response{data=model.LlmopsAgentEval} "评测完成"
// @Router /llmops/agent/eval/run [post]
func (a *LlmopsApi) RunAgentEval(c *gin.Context) {
	var req idReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	ev, err := service.LlmopsService.RunAgentEval(req.ID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(ev, "评测完成", c)
}

// GetAgentEvalResults 智能体评测明细
// @Tags Llmops
// @Summary 智能体评测明细
// @Security ApiKeyAuth
// @Param evalId query int true "评测ID"
// @Success 200 {object} response.Response{data=[]model.LlmopsAgentEvalResult} "获取成功"
// @Router /llmops/agent/eval/results [get]
func (a *LlmopsApi) GetAgentEvalResults(c *gin.Context) {
	evalId, err := strconv.ParseUint(c.Query("evalId"), 10, 64)
	if err != nil || evalId == 0 {
		response.FailWithMessage("evalId 不能为空", c)
		return
	}
	list, err := service.LlmopsService.GetAgentEvalResults(uint(evalId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// RollbackAgent 回滚版本
// @Tags Llmops
// @Summary 智能体回滚到指定版本（快照回写）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.AgentRollbackIn true "回滚请求"
// @Success 200 {object} response.Response{data=model.LlmopsAgent} "回滚成功"
// @Router /llmops/agent/rollback [post]
func (a *LlmopsApi) RollbackAgent(c *gin.Context) {
	var in service.AgentRollbackIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	ag, err := service.LlmopsService.RollbackAgent(in.AgentId, in.Version)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(ag, "回滚成功", c)
}

// GetAgentCallStats 调用统计
// @Tags Llmops
// @Summary 智能体调用统计（次数/成功失败/平均与 P95 时延，近 N 小时）
// @Security ApiKeyAuth
// @Param agentId query int true "智能体ID"
// @Param hours query int false "统计窗口小时数(默认 24)"
// @Success 200 {object} response.Response{data=service.AgentCallStat} "获取成功"
// @Router /llmops/agent/stats [get]
func (a *LlmopsApi) GetAgentCallStats(c *gin.Context) {
	agentId, err := strconv.ParseUint(c.Query("agentId"), 10, 64)
	if err != nil || agentId == 0 {
		response.FailWithMessage("agentId 不能为空", c)
		return
	}
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))
	stat, err := service.LlmopsService.GetAgentCallStats(uint(agentId), hours)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(stat, "获取成功", c)
}

// GetAgentMonitor 监控聚合
// @Tags Llmops
// @Summary 智能体运行监控（统计+时延四桶分布+异常会话数）
// @Security ApiKeyAuth
// @Param agentId query int true "智能体ID"
// @Param hours query int false "窗口小时数(默认 24)"
// @Success 200 {object} response.Response{data=service.AgentMonitorView} "获取成功"
// @Router /llmops/agent/monitor [get]
func (a *LlmopsApi) GetAgentMonitor(c *gin.Context) {
	agentId, err := strconv.ParseUint(c.Query("agentId"), 10, 64)
	if err != nil || agentId == 0 {
		response.FailWithMessage("agentId 不能为空", c)
		return
	}
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))
	view, err := service.LlmopsService.GetAgentMonitor(uint(agentId), hours)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(view, "获取成功", c)
}

// GetSkills 技能列表
// @Tags Llmops
// @Summary 技能列表（类型/状态过滤）
// @Security ApiKeyAuth
// @Param type query string false "builtin/extended/custom"
// @Param status query string false "enabled/disabled"
// @Success 200 {object} response.Response{data=[]model.LlmopsSkill} "获取成功"
// @Router /llmops/skill/list [get]
func (a *LlmopsApi) GetSkills(c *gin.Context) {
	list, err := service.LlmopsService.GetSkills(c.Query("type"), c.Query("status"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveSkill 保存技能
// @Tags Llmops
// @Summary 新建/更新技能（定义即检，prompt 型协议）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.LlmopsSkill true "技能"
// @Success 200 {object} response.Response{data=model.LlmopsSkill} "保存成功"
// @Router /llmops/skill [post]
func (a *LlmopsApi) SaveSkill(c *gin.Context) {
	var in model.LlmopsSkill
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	sk, err := service.LlmopsService.SaveSkill(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(sk, "保存成功", c)
}

// DeleteSkill 删除技能
// @Tags Llmops
// @Summary 删除技能（软删除）
// @Security ApiKeyAuth
// @Param id query int true "技能ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/skill [delete]
func (a *LlmopsApi) DeleteSkill(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteSkill(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GenerateSkill 零代码创建技能
// @Tags Llmops
// @Summary 自然语言零代码创建技能（LLM 生成定义，校验后 custom 入库）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.SkillGenIn true "生成请求"
// @Success 200 {object} response.Response{data=model.LlmopsSkill} "生成成功"
// @Router /llmops/skill/generate [post]
func (a *LlmopsApi) GenerateSkill(c *gin.Context) {
	var in service.SkillGenIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	sk, err := service.LlmopsService.CreateSkillFromDescription(in.Desc)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(sk, "生成成功", c)
}

// ExportSkillPackage 技能包导出
// @Tags Llmops
// @Summary 按 ID 列表导出技能包 JSON
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.SkillExportIn true "导出请求"
// @Success 200 {object} response.Response{data=string} "导出成功"
// @Router /llmops/skill/export [post]
func (a *LlmopsApi) ExportSkillPackage(c *gin.Context) {
	var in service.SkillExportIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	pkg, err := service.LlmopsService.ExportSkillPackage(in.Ids)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(pkg, "导出成功", c)
}

// ImportSkillPackage 技能包导入
// @Tags Llmops
// @Summary 一键安装技能包（默认冲突跳过并报告，overwrite 覆盖）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.SkillImportIn true "导入请求"
// @Success 200 {object} response.Response{data=map[string]any} "导入完成"
// @Router /llmops/skill/import [post]
func (a *LlmopsApi) ImportSkillPackage(c *gin.Context) {
	var in service.SkillImportIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	imported, conflicts, err := service.LlmopsService.ImportSkillPackage(in.Raw, in.Overwrite)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"imported": imported, "conflicts": conflicts}, "导入完成", c)
}

// GetMcpTools MCP 工具列表
// @Tags Llmops
// @Summary MCP 工具列表（固定头不回显）
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]model.LlmopsMcpTool} "获取成功"
// @Router /llmops/mcp/list [get]
func (a *LlmopsApi) GetMcpTools(c *gin.Context) {
	list, err := service.LlmopsService.GetMcpTools()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveMcpTool 保存 MCP 工具
// @Tags Llmops
// @Summary 新建/更新 MCP 工具（headersJSON 明文入参，逐值加密不回显）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.McpToolIn true "MCP 工具"
// @Success 200 {object} response.Response{data=model.LlmopsMcpTool} "保存成功"
// @Router /llmops/mcp [post]
func (a *LlmopsApi) SaveMcpTool(c *gin.Context) {
	var in service.McpToolIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	t, err := service.LlmopsService.SaveMcpTool(&in.LlmopsMcpTool, in.HeadersJSON)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(t, "保存成功", c)
}

// DeleteMcpTool 删除 MCP 工具
// @Tags Llmops
// @Summary 删除 MCP 工具（软删除）
// @Security ApiKeyAuth
// @Param id query int true "工具ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/mcp [delete]
func (a *LlmopsApi) DeleteMcpTool(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteMcpTool(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetConnectors 连接器列表
// @Tags Llmops
// @Summary 业务系统连接器列表（鉴权值不回显）
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]model.LlmopsConnector} "获取成功"
// @Router /llmops/connector/list [get]
func (a *LlmopsApi) GetConnectors(c *gin.Context) {
	list, err := service.LlmopsService.GetConnectors()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveConnector 保存连接器
// @Tags Llmops
// @Summary 新建/更新连接器（authValue 留空不改，密文落库）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.ConnectorIn true "连接器"
// @Success 200 {object} response.Response{data=model.LlmopsConnector} "保存成功"
// @Router /llmops/connector [post]
func (a *LlmopsApi) SaveConnector(c *gin.Context) {
	var in service.ConnectorIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	conn, err := service.LlmopsService.SaveConnector(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(conn, "保存成功", c)
}

// TestConnector 连通测试
// @Tags Llmops
// @Summary 连接器连通测试（带鉴权头探测）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body idReq true "连接器ID"
// @Success 200 {object} response.Response{msg=string} "连通正常"
// @Router /llmops/connector/test [post]
func (a *LlmopsApi) TestConnector(c *gin.Context) {
	var req idReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.LlmopsService.TestConnector(req.ID); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("连通正常", c)
}

// DeleteConnector 删除连接器
// @Tags Llmops
// @Summary 删除连接器（软删除）
// @Security ApiKeyAuth
// @Param id query int true "连接器ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/connector [delete]
func (a *LlmopsApi) DeleteConnector(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteConnector(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetSkillTokens 令牌列表
// @Tags Llmops
// @Summary 技能调用令牌列表（哈希不回显）
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]model.LlmopsSkillToken} "获取成功"
// @Router /llmops/skill/token/list [get]
func (a *LlmopsApi) GetSkillTokens(c *gin.Context) {
	list, err := service.LlmopsService.GetSkillTokens()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// CreateSkillToken 创建令牌
// @Tags Llmops
// @Summary 创建调用令牌（明文仅本次返回）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.TokenIn true "令牌"
// @Success 200 {object} response.Response{data=map[string]any} "创建成功"
// @Router /llmops/skill/token [post]
func (a *LlmopsApi) CreateSkillToken(c *gin.Context) {
	var in service.TokenIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	t, plain, err := service.LlmopsService.CreateSkillToken(in.Name, in.RateLimit)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"token": t, "plain": plain}, "创建成功（明文仅此一次，请妥善保存）", c)
}

// RevokeSkillToken 吊销令牌
// @Tags Llmops
// @Summary 吊销调用令牌
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body idReq true "令牌ID"
// @Success 200 {object} response.Response{msg=string} "已吊销"
// @Router /llmops/skill/token/revoke [post]
func (a *LlmopsApi) RevokeSkillToken(c *gin.Context) {
	var req idReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.LlmopsService.RevokeSkillToken(req.ID); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("已吊销", c)
}

// SkillCall 开放调用入口
// @Tags Llmops
// @Summary 技能开放调用（令牌+限流+沙箱+审计）
// @Accept application/json
// @Param data body service.SkillCallIn true "调用请求"
// @Success 200 {object} response.Response{data=string} "调用成功"
// @Router /llmops/skill/call [post]
func (a *LlmopsApi) SkillCall(c *gin.Context) {
	var in service.SkillCallIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	result, err := service.LlmopsService.SkillCallWithAuth(in.Token, in.Tool, in.Args)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(result, "调用成功", c)
}

// GetSkillCallLogs 审计日志
// @Tags Llmops
// @Summary 技能调用审计日志（令牌过滤，倒序分页）
// @Security ApiKeyAuth
// @Param tokenId query int false "令牌ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /llmops/skill/callLogs [get]
func (a *LlmopsApi) GetSkillCallLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	tokenId, _ := strconv.ParseUint(c.Query("tokenId"), 10, 64)
	list, total, err := service.LlmopsService.GetSkillCallLogs(page, pageSize, uint(tokenId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetPrompts 模板列表
// @Tags Llmops
// @Summary Prompt 模板列表
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]model.LlmopsPromptTemplate} "获取成功"
// @Router /llmops/prompt/list [get]
func (a *LlmopsApi) GetPrompts(c *gin.Context) {
	list, err := service.LlmopsService.GetPrompts()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SavePrompt 保存模板
// @Tags Llmops
// @Summary 新建/更新 Prompt 模板（更新版本自增+旧版快照）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.LlmopsPromptTemplate true "模板"
// @Success 200 {object} response.Response{data=model.LlmopsPromptTemplate} "保存成功"
// @Router /llmops/prompt [post]
func (a *LlmopsApi) SavePrompt(c *gin.Context) {
	var in model.LlmopsPromptTemplate
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	tpl, err := service.LlmopsService.SavePrompt(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(tpl, "保存成功", c)
}

// DeletePrompt 删除模板
// @Tags Llmops
// @Summary 删除 Prompt 模板（级联删版本快照）
// @Security ApiKeyAuth
// @Param id query int true "模板ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/prompt [delete]
func (a *LlmopsApi) DeletePrompt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeletePrompt(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetPromptVersions 模板版本列表
// @Tags Llmops
// @Summary Prompt 模板版本快照列表
// @Security ApiKeyAuth
// @Param templateId query int true "模板ID"
// @Success 200 {object} response.Response{data=[]model.LlmopsPromptVersion} "获取成功"
// @Router /llmops/prompt/versions [get]
func (a *LlmopsApi) GetPromptVersions(c *gin.Context) {
	templateId, err := strconv.ParseUint(c.Query("templateId"), 10, 64)
	if err != nil || templateId == 0 {
		response.FailWithMessage("templateId 不能为空", c)
		return
	}
	list, err := service.LlmopsService.GetPromptVersions(uint(templateId))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// RenderPrompt 模板渲染
// @Tags Llmops
// @Summary Prompt 模板渲染（templateId 或内联 content，缺失变量报错列清单）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.RenderPromptIn true "渲染请求"
// @Success 200 {object} response.Response{data=map[string]any} "渲染完成"
// @Router /llmops/prompt/render [post]
func (a *LlmopsApi) RenderPrompt(c *gin.Context) {
	var in service.RenderPromptIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	rendered, usedVars, err := service.LlmopsService.RenderPrompt(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"rendered": rendered, "vars": usedVars}, "渲染完成", c)
}

// OptimizePrompts Prompt 优化
// @Tags Llmops
// @Summary Prompt 在线/批量优化（≤10 条，逐条结果带原版对比，单条失败不中断）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.OptimizeIn true "优化请求"
// @Success 200 {object} response.Response{data=[]service.OptimizeResult} "优化完成"
// @Router /llmops/prompt/optimize [post]
func (a *LlmopsApi) OptimizePrompts(c *gin.Context) {
	var in service.OptimizeIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	results, err := service.LlmopsService.OptimizePrompts(in.Items, in.ModelId)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(results, "优化完成", c)
}

// GetPromptAssets 配套资产列表
// @Tags Llmops
// @Summary Prompt 配套资产列表（业务规则/术语/思维链，类型过滤）
// @Security ApiKeyAuth
// @Param kind query string false "rule/term/cot"
// @Success 200 {object} response.Response{data=[]model.LlmopsPromptAsset} "获取成功"
// @Router /llmops/prompt/asset/list [get]
func (a *LlmopsApi) GetPromptAssets(c *gin.Context) {
	list, err := service.LlmopsService.GetPromptAssets(c.Query("kind"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SavePromptAsset 保存配套资产
// @Tags Llmops
// @Summary 新建/更新 Prompt 配套资产
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.LlmopsPromptAsset true "资产"
// @Success 200 {object} response.Response{data=model.LlmopsPromptAsset} "保存成功"
// @Router /llmops/prompt/asset [post]
func (a *LlmopsApi) SavePromptAsset(c *gin.Context) {
	var in model.LlmopsPromptAsset
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	asset, err := service.LlmopsService.SavePromptAsset(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(asset, "保存成功", c)
}

// DeletePromptAsset 删除配套资产
// @Tags Llmops
// @Summary 删除 Prompt 配套资产
// @Security ApiKeyAuth
// @Param id query int true "资产ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/prompt/asset [delete]
func (a *LlmopsApi) DeletePromptAsset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeletePromptAsset(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetAgentAbnormalSessions 异常会话列表
// @Tags Llmops
// @Summary 异常会话列表（窗口内存在失败调用的会话，按失败次数降序）
// @Security ApiKeyAuth
// @Param agentId query int true "智能体ID"
// @Param hours query int false "窗口小时数(默认 24)"
// @Param page query int false "页码"
// @Param pageSize query int false "每页条数"
// @Success 200 {object} response.Response{data=map[string]any} "获取成功"
// @Router /llmops/agent/abnormalSessions [get]
func (a *LlmopsApi) GetAgentAbnormalSessions(c *gin.Context) {
	agentId, err := strconv.ParseUint(c.Query("agentId"), 10, 64)
	if err != nil || agentId == 0 {
		response.FailWithMessage("agentId 不能为空", c)
		return
	}
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := service.LlmopsService.GetAgentAbnormalSessions(uint(agentId), hours, page, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "获取成功", c)
}

// GetDatasets 数据集列表
// @Tags Llmops
// @Summary 数据集列表（用途过滤）
// @Security ApiKeyAuth
// @Param purpose query string false "sft/rlhf/eval/exam"
// @Success 200 {object} response.Response{data=[]model.LlmopsDataset} "获取成功"
// @Router /llmops/dataset/list [get]
func (a *LlmopsApi) GetDatasets(c *gin.Context) {
	list, err := service.LlmopsService.GetDatasets(c.Query("purpose"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveDataset 保存数据集
// @Tags Llmops
// @Summary 新建/更新数据集（样本数自动计数）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.LlmopsDataset true "数据集"
// @Success 200 {object} response.Response{data=model.LlmopsDataset} "保存成功"
// @Router /llmops/dataset [post]
func (a *LlmopsApi) SaveDataset(c *gin.Context) {
	var in model.LlmopsDataset
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	d, err := service.LlmopsService.SaveDataset(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(d, "保存成功", c)
}

// DeleteDataset 删除数据集
// @Tags Llmops
// @Summary 删除数据集（软删除）
// @Security ApiKeyAuth
// @Param id query int true "数据集ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/dataset [delete]
func (a *LlmopsApi) DeleteDataset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteDataset(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ImportDatasetContent 数据集导入
// @Tags Llmops
// @Summary 数据集内容导入（JSONL 逐行校验，仅有效行入库，坏行报告）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.DatasetImportIn true "导入请求"
// @Success 200 {object} response.Response{data=map[string]any} "导入完成"
// @Router /llmops/dataset/import [post]
func (a *LlmopsApi) ImportDatasetContent(c *gin.Context) {
	var in service.DatasetImportIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	valid, errs, err := service.LlmopsService.ImportDatasetContent(in.ID, in.Content, in.Purpose)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"valid": valid, "errors": errs}, "导入完成", c)
}

// ExportDataset 数据集导出
// @Tags Llmops
// @Summary 数据集内容导出（原样返回供下载）
// @Security ApiKeyAuth
// @Param id query int true "数据集ID"
// @Success 200 {object} response.Response{data=map[string]any} "导出成功"
// @Router /llmops/dataset/export [get]
func (a *LlmopsApi) ExportDataset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	content, filename, err := service.LlmopsService.ExportDataset(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"content": content, "filename": filename}, "导出成功", c)
}

// GetDatasetRefs 数据集引用计数
// @Tags Llmops
// @Summary 数据集被任务引用计数（训练任务/评估任务）
// @Security ApiKeyAuth
// @Param id query int true "数据集ID"
// @Success 200 {object} response.Response{data=service.DatasetRefs} "获取成功"
// @Router /llmops/dataset/refs [get]
func (a *LlmopsApi) GetDatasetRefs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	refs, err := service.LlmopsService.GetDatasetRefs(uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(refs, "获取成功", c)
}

// GetDashboard 数据看板
// @Tags Llmops
// @Summary 平台数据看板（各域计数+近 7 日调用趋势）
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=service.DashboardStats} "获取成功"
// @Router /llmops/dashboard [get]
func (a *LlmopsApi) GetDashboard(c *gin.Context) {
	st, err := service.LlmopsService.GetDashboard()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(st, "获取成功", c)
}

// GetQuotas 配额列表
// @Tags Llmops
// @Summary 资源配额列表（维度过滤）
// @Security ApiKeyAuth
// @Param scope query string false "user/agent"
// @Success 200 {object} response.Response{data=[]model.LlmopsQuota} "获取成功"
// @Router /llmops/quota/list [get]
func (a *LlmopsApi) GetQuotas(c *gin.Context) {
	list, err := service.LlmopsService.GetQuotas(c.Query("scope"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SetQuota 设置配额
// @Tags Llmops
// @Summary 设置资源配额（upsert；-1 表示不限）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.QuotaIn true "配额"
// @Success 200 {object} response.Response{data=model.LlmopsQuota} "设置成功"
// @Router /llmops/quota [post]
func (a *LlmopsApi) SetQuota(c *gin.Context) {
	var in service.QuotaIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	q, err := service.LlmopsService.SetQuota(in.Scope, in.ScopeKey, in.DailyTokenLimit, in.DailyCallLimit)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(q, "设置成功", c)
}

// ConsumeQuota 配额扣减
// @Tags Llmops
// @Summary 配额扣减（行锁事务：跨日重置，超限拒）
// @Accept application/json
// @Param data body service.QuotaConsumeIn true "扣减请求"
// @Success 200 {object} response.Response{data=model.LlmopsQuota} "扣减成功"
// @Router /llmops/quota/consume [post]
func (a *LlmopsApi) ConsumeQuota(c *gin.Context) {
	var in service.QuotaConsumeIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	q, err := service.LlmopsService.ConsumeQuota(in.Scope, in.ScopeKey, in.Tokens, in.Calls)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(q, "扣减成功", c)
}

// TakeOverSession 人工接管
// @Tags Llmops
// @Summary 会话人工接管（接管中拒自动回复）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body service.TakeOverIn true "接管请求"
// @Success 200 {object} response.Response{data=model.LlmopsAgentSession} "接管成功"
// @Router /llmops/intervene/takeover [post]
func (a *LlmopsApi) TakeOverSession(c *gin.Context) {
	var in service.TakeOverIn
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	sess, err := service.LlmopsService.TakeOverSession(in.SessionId, in.Operator)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(sess, "接管成功", c)
}

// ReleaseSession 解除接管
// @Tags Llmops
// @Summary 解除人工接管
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body idReq true "会话ID"
// @Success 200 {object} response.Response{data=model.LlmopsAgentSession} "已解除"
// @Router /llmops/intervene/release [post]
func (a *LlmopsApi) ReleaseSession(c *gin.Context) {
	var req idReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	sess, err := service.LlmopsService.ReleaseSession(req.ID)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(sess, "已解除", c)
}

// GetSensitiveWords 敏感词列表
// @Tags Llmops
// @Summary 敏感词列表
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]model.LlmopsSensitiveWord} "获取成功"
// @Router /llmops/sensitive/list [get]
func (a *LlmopsApi) GetSensitiveWords(c *gin.Context) {
	list, err := service.LlmopsService.ListSensitiveWords()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// SaveSensitiveWord 保存敏感词
// @Tags Llmops
// @Summary 新增/更新敏感词（block/replace）
// @Security ApiKeyAuth
// @Accept application/json
// @Param data body model.LlmopsSensitiveWord true "敏感词"
// @Success 200 {object} response.Response{data=model.LlmopsSensitiveWord} "保存成功"
// @Router /llmops/sensitive [post]
func (a *LlmopsApi) SaveSensitiveWord(c *gin.Context) {
	var in model.LlmopsSensitiveWord
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	w, err := service.LlmopsService.SaveSensitiveWord(&in)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(w, "保存成功", c)
}

// DeleteSensitiveWord 删除敏感词
// @Tags Llmops
// @Summary 删除敏感词
// @Security ApiKeyAuth
// @Param id query int true "敏感词ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /llmops/sensitive [delete]
func (a *LlmopsApi) DeleteSensitiveWord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("id 不能为空", c)
		return
	}
	if err := service.LlmopsService.DeleteSensitiveWord(uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
