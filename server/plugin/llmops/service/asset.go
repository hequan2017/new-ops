// Package service llmops：数据资产管理（列表/详情/软删除，采集解析随后续条目扩展）
package service

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段 2200-2299）
const (
	ErrCodeAssetNotFound = 2200
	ErrCodeAssetTypeBad  = 2201
)

// DataAssetTypes 数据类型白名单（要求(1)：文档/图片/音频/视频/网页/问答对 + 结构化/半结构化）
var DataAssetTypes = []string{"doc", "image", "audio", "video", "webpage", "qa", "structured", "semistructured"}

// DataAssetStatuses 资产生命周期状态白名单
var DataAssetStatuses = []string{"draft", "parsed", "cataloged", "offline"}

var LlmopsService = new(LlmopsServiceStruct)

type LlmopsServiceStruct struct{}

// ValidDataType 数据类型是否在白名单内
func ValidDataType(t string) bool {
	for _, v := range DataAssetTypes {
		if v == t {
			return true
		}
	}
	return false
}

// ValidAssetStatus 生命周期状态是否在白名单内
func ValidAssetStatus(s string) bool {
	for _, v := range DataAssetStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// EscapeLike 转义 LIKE 通配符，保证关键字按字面匹配
func EscapeLike(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

// GetAssetList 数据资产列表（分页 + 类型/状态/关键字过滤，全参数白名单校验）
func (s *LlmopsServiceStruct) GetAssetList(page, pageSize int, dataType, status, keyword string) ([]model.LlmopsDataAsset, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	db := global.GVA_DB.Model(&model.LlmopsDataAsset{})
	if dataType != "" {
		if !ValidDataType(dataType) {
			return nil, 0, fmt.Errorf("[%d]不支持的数据类型: %s", ErrCodeAssetTypeBad, dataType)
		}
		db = db.Where("data_type = ?", dataType)
	}
	if status != "" {
		if !ValidAssetStatus(status) {
			return nil, 0, fmt.Errorf("[%d]不支持的状态: %s", ErrCodeAssetTypeBad, status)
		}
		db = db.Where("status = ?", status)
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		like := "%" + EscapeLike(kw) + "%"
		db = db.Where("name LIKE ? OR tags LIKE ?", like, like)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.LlmopsDataAsset
	err := db.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// GetAsset 数据资产详情
func (s *LlmopsServiceStruct) GetAsset(id uint) (*model.LlmopsDataAsset, error) {
	var asset model.LlmopsDataAsset
	err := global.GVA_DB.First(&asset, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("[%d]数据资产不存在", ErrCodeAssetNotFound)
		}
		return nil, err
	}
	return &asset, nil
}

// DeleteAsset 数据资产软删除（删除类接口软删除红线）
func (s *LlmopsServiceStruct) DeleteAsset(id uint) error {
	res := global.GVA_DB.Delete(&model.LlmopsDataAsset{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]数据资产不存在", ErrCodeAssetNotFound)
	}
	return nil
}
