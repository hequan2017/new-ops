// Package service llmops：Prompt 配套资产（要求(7)）——业务规则/行业术语库/专家思维链统一管理
// 口径：三类资产统一表 kind 区分（rule/term/cot）；可挂引用模板（联动展示）。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const ErrCodePromptAssetKind = 2298

// PromptAssetKinds 资产类型白名单
var PromptAssetKinds = []string{"rule", "term", "cot"}

// ValidPromptAssetKind 类型白名单
func ValidPromptAssetKind(k string) bool {
	for _, v := range PromptAssetKinds {
		if v == k {
			return true
		}
	}
	return false
}

// GetPromptAssets 资产列表（类型过滤）
func (s *LlmopsServiceStruct) GetPromptAssets(kind string) ([]model.LlmopsPromptAsset, error) {
	db := global.GVA_DB.Model(&model.LlmopsPromptAsset{})
	if kind != "" {
		if !ValidPromptAssetKind(kind) {
			return nil, fmt.Errorf("[%d]不支持的资产类型: %s", ErrCodePromptAssetKind, kind)
		}
		db = db.Where("kind = ?", kind)
	}
	var list []model.LlmopsPromptAsset
	err := db.Order("id desc").Find(&list).Error
	return list, err
}

// SavePromptAsset 新建/更新资产
func (s *LlmopsServiceStruct) SavePromptAsset(in *model.LlmopsPromptAsset) (*model.LlmopsPromptAsset, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Content) == "" || !ValidPromptAssetKind(in.Kind) {
		return nil, fmt.Errorf("[%d]名称/内容必填且类型须为 rule/term/cot", ErrCodePromptAssetKind)
	}
	if in.ID > 0 {
		var a model.LlmopsPromptAsset
		if err := global.GVA_DB.First(&a, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]资产不存在", ErrCodePromptNotFound)
		}
		err := global.GVA_DB.Model(&a).Updates(map[string]any{
			"name": in.Name, "kind": in.Kind, "content": in.Content, "template_id": in.TemplateId,
		}).Error
		if err != nil {
			return nil, err
		}
		return &a, nil
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeletePromptAsset 删除资产
func (s *LlmopsServiceStruct) DeletePromptAsset(id uint) error {
	res := global.GVA_DB.Delete(&model.LlmopsPromptAsset{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]资产不存在", ErrCodePromptNotFound)
	}
	return nil
}
