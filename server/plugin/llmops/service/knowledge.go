// Package service llmops：知识库管理（要求(2) 四类知识库 CRUD + 文档入库）
package service

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodeKbNotFound  = 2238
	ErrCodeKbTypeBad   = 2239
	ErrCodeKbDocMiss   = 2240
	ErrCodeKbEmpty     = 2241
)

// KnowledgeTypes 四类知识库白名单
var KnowledgeTypes = []string{"vector", "fulltext", "graph", "qa"}

// ValidKnowledgeType 知识库类型白名单
func ValidKnowledgeType(t string) bool {
	for _, v := range KnowledgeTypes {
		if v == t {
			return true
		}
	}
	return false
}

// KnowledgeIn 知识库保存入参
type KnowledgeIn struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Desc         string `json:"desc"`
	Status       string `json:"status"`
	EmbedModelId uint   `json:"embedModelId"`
}

// IngestIn 知识文档入库请求（assetId 与 text 二选一，优先资产）
type IngestIn struct {
	KbId    uint   `json:"kbId" binding:"required"`
	AssetId uint   `json:"assetId"`
	Title   string `json:"title"`
	Text    string `json:"text"`
}

// VectorIndexIn 向量索引请求
type VectorIndexIn struct {
	KbId  uint `json:"kbId" binding:"required"`
	DocId uint `json:"docId" binding:"required"`
}

// VectorSearchIn 向量检索请求
type VectorSearchIn struct {
	KbId  uint   `json:"kbId" binding:"required"`
	Query string `json:"query" binding:"required"`
	TopK  int    `json:"topK"`
}

// QaSearchIn 问答检索请求
type QaSearchIn struct {
	KbId     uint    `json:"kbId" binding:"required"`
	Query    string  `json:"query" binding:"required"`
	TopK     int     `json:"topK"`
	MinScore float64 `json:"minScore"`
}

// GraphQueryIn 图谱多跳推理请求
type GraphQueryIn struct {
	KbId        uint   `json:"kbId" binding:"required"`
	StartNodeId uint   `json:"startNodeId" binding:"required"`
	MaxHops     int    `json:"maxHops"`
	Relation    string `json:"relation"`
}

// SummaryGenIn 摘要生成请求
type SummaryGenIn struct {
	KbId    uint `json:"kbId" binding:"required"`
	DocId   uint `json:"docId" binding:"required"`
	ModelId uint `json:"modelId" binding:"required"`
}

// TrainTransitionIn 训练任务流转请求
type TrainTransitionIn struct {
	ID     uint   `json:"id" binding:"required"`
	To     string `json:"to" binding:"required"`
	Reason string `json:"reason"`
}

// TrainStopIn 训练任务停止请求
type TrainStopIn struct {
	ID     uint   `json:"id" binding:"required"`
	Reason string `json:"reason"`
}

// AgentStageIn 智能体阶段流转请求
type AgentStageIn struct {
	ID    uint   `json:"id" binding:"required"`
	Stage string `json:"stage" binding:"required"`
}

// AgentPublishIn 智能体发布请求
type AgentPublishIn struct {
	ID     uint   `json:"id" binding:"required"`
	Remark string `json:"remark"`
}

// GetKnowledges 知识库列表
func (s *LlmopsServiceStruct) GetKnowledges() ([]model.LlmopsKnowledge, error) {
	var list []model.LlmopsKnowledge
	err := global.GVA_DB.Order("id").Find(&list).Error
	return list, err
}

// SaveKnowledge 新建/更新知识库（类型白名单；vector 库可绑默认嵌入模型）
func (s *LlmopsServiceStruct) SaveKnowledge(in *KnowledgeIn) (*model.LlmopsKnowledge, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || !ValidKnowledgeType(in.Type) {
		return nil, fmt.Errorf("[%d]名称必填且类型须为 vector/fulltext/graph/qa", ErrCodeKbTypeBad)
	}
	if in.Status == "" {
		in.Status = "active"
	}
	var count int64
	q := global.GVA_DB.Model(&model.LlmopsKnowledge{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]知识库名已存在", ErrCodeKbTypeBad)
	}
	if in.ID > 0 {
		var kb model.LlmopsKnowledge
		if err := global.GVA_DB.First(&kb, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]知识库不存在", ErrCodeKbNotFound)
		}
		err := global.GVA_DB.Model(&kb).Updates(map[string]any{
			"name": in.Name, "type": in.Type, "desc": in.Desc,
			"status": in.Status, "embed_model_id": in.EmbedModelId,
		}).Error
		if err != nil {
			return nil, err
		}
		return &kb, nil
	}
	kb := &model.LlmopsKnowledge{
		Name: in.Name, Type: in.Type, Desc: in.Desc,
		Status: in.Status, EmbedModelId: in.EmbedModelId,
	}
	if err := global.GVA_DB.Create(kb).Error; err != nil {
		return nil, err
	}
	return kb, nil
}

// DeleteKnowledge 删除知识库（软删除；文档级联软删）
func (s *LlmopsServiceStruct) DeleteKnowledge(id uint) error {
	var kb model.LlmopsKnowledge
	if err := global.GVA_DB.First(&kb, id).Error; err != nil {
		return fmt.Errorf("[%d]知识库不存在", ErrCodeKbNotFound)
	}
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&kb).Error; err != nil {
			return err
		}
		return tx.Where("kb_id = ?", id).Delete(&model.LlmopsKnowledgeDoc{}).Error
	})
}

// IngestDocument 文档入库：来源资产(rawText)或内联文本，原文留档为 doc 行
func (s *LlmopsServiceStruct) IngestDocument(kbId uint, assetId uint, title, text string) (*model.LlmopsKnowledgeDoc, error) {
	var kb model.LlmopsKnowledge
	if err := global.GVA_DB.First(&kb, kbId).Error; err != nil {
		return nil, fmt.Errorf("[%d]知识库不存在", ErrCodeKbNotFound)
	}
	content := strings.TrimSpace(text)
	if assetId > 0 {
		raw, err := s.GetAssetDocContent(assetId)
		if err != nil {
			return nil, err
		}
		content = raw
		if title == "" {
			var asset model.LlmopsDataAsset
			if err := global.GVA_DB.First(&asset, assetId).Error; err == nil {
				title = asset.Name
			}
		}
	}
	if content == "" {
		return nil, fmt.Errorf("[%d]入库内容为空（资产无正文或未提供文本）", ErrCodeKbEmpty)
	}
	if strings.TrimSpace(title) == "" {
		title = "未命名文档"
	}
	doc := &model.LlmopsKnowledgeDoc{
		KbId: kbId, AssetId: assetId, Title: strings.TrimSpace(title),
		Content: content, Status: "ready",
	}
	if err := global.GVA_DB.Create(doc).Error; err != nil {
		return nil, err
	}
	return doc, nil
}

// GetKnowledgeDocs 库内文档列表（分页，不含正文大字段）
func (s *LlmopsServiceStruct) GetKnowledgeDocs(kbId uint, page, pageSize int) ([]model.LlmopsKnowledgeDoc, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	db := global.GVA_DB.Model(&model.LlmopsKnowledgeDoc{}).Select("id, created_at, updated_at, deleted_at, kb_id, asset_id, title, status, chunk_count").
		Where("kb_id = ?", kbId)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.LlmopsKnowledgeDoc
	err := db.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// DeleteKnowledgeDoc 删除库内文档（软删除）
func (s *LlmopsServiceStruct) DeleteKnowledgeDoc(id uint) error {
	res := global.GVA_DB.Delete(&model.LlmopsKnowledgeDoc{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]文档不存在", ErrCodeKbDocMiss)
	}
	return nil
}
