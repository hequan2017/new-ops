// Package service crossauth：角色模板（需求 3.7-5）
// 口径：内置基础权限 JSON 结构校验（纯函数：须为对象且含 grants 数组）；
// 模板 8 条种子按角色名幂等；查询 API（分类过滤）。
package service

import (
	"encoding/json"
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// 错误码（crossauth 段 2900-2999）
const ErrCodeCaRoleParam = 2908

// validatePermJSON 权限 JSON 校验（纯函数，测试锚点）：对象且显式含 grants 数组
func validatePermJSON(s string) error {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal([]byte(s), &probe); err != nil {
		return fmt.Errorf("[%d]权限须为 JSON 对象: %v", ErrCodeCaRoleParam, err)
	}
	raw, ok := probe["grants"]
	if !ok {
		return fmt.Errorf("[%d]权限 JSON 须显式含 grants 数组", ErrCodeCaRoleParam)
	}
	var grants []struct {
		Resource string   `json:"resource"`
		Actions  []string `json:"actions"`
	}
	if err := json.Unmarshal(raw, &grants); err != nil {
		return fmt.Errorf("[%d]grants 须为数组: %v", ErrCodeCaRoleParam, err)
	}
	return nil
}

// SaveCaRoleTemplate 新建/更新模板（权限 JSON 校验）
func (s *CaService) SaveCaRoleTemplate(in *model.CaRoleTemplate) (*model.CaRoleTemplate, error) {
	if err := validatePermJSON(in.PermJSON); err != nil {
		return nil, err
	}
	var exist model.CaRoleTemplate
	if err := global.GVA_DB.Where("role_name = ?", in.RoleName).First(&exist).Error; err == nil {
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"category": in.Category, "perm_json": in.PermJSON, "description": in.Description,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetCaRoleTemplates 模板列表（分类过滤）
func (s *CaService) GetCaRoleTemplates(category string) ([]model.CaRoleTemplate, error) {
	db := global.GVA_DB.Model(&model.CaRoleTemplate{})
	if category != "" {
		db = db.Where("category = ?", category)
	}
	var list []model.CaRoleTemplate
	err := db.Order("id ASC").Limit(100).Find(&list).Error
	return list, err
}
