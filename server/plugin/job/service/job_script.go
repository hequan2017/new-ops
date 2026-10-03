// Package service job 插件：脚本库与变量组服务
package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/job/model"
	"gorm.io/gorm"
)

// 脚本库/变量组错误码（批量执行见 errors.go 1211-1214）
const (
	ErrCodeScriptNameRequired = 1201
	ErrCodeScriptDuplicate    = 1202
	ErrCodeScriptLangInvalid  = 1203
	ErrCodeScriptNotFound     = 1204
	ErrCodeVarGroupRequired   = 1205
	ErrCodeVarGroupDuplicate  = 1206
	ErrCodeVariablesInvalid   = 1207
)

// ScriptService 脚本库服务
type ScriptService struct{}

// validateScript 脚本校验（纯逻辑）
func validateScript(sc *model.JobScript) error {
	if sc == nil || sc.Name == "" {
		return newScriptErr(ErrCodeScriptNameRequired, "脚本名称不能为空")
	}
	if !model.ValidScriptLangs()[sc.Language] {
		return newScriptErr(ErrCodeScriptLangInvalid, fmt.Sprintf("非法的脚本语言: %s", sc.Language))
	}
	return nil
}

// CreateScript 创建脚本（版本号 1）
func (s *ScriptService) CreateScript(sc *model.JobScript, operator string) error {
	if err := validateScript(sc); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.JobScript{}).Where("name = ?", sc.Name).Count(&count)
	if count > 0 {
		return newScriptErr(ErrCodeScriptDuplicate, fmt.Sprintf("脚本已存在: %s", sc.Name))
	}
	sc.Version = 1
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(sc).Error; err != nil {
			return err
		}
		return tx.Create(&model.JobScriptVersion{ScriptID: sc.ID, Version: 1, Content: sc.Content, Operator: operator}).Error
	})
}

// UpdateScript 更新脚本（内容变更时版本递增并归档旧内容）
func (s *ScriptService) UpdateScript(sc *model.JobScript, operator string) error {
	if sc == nil || sc.ID == 0 || sc.Name == "" {
		return newScriptErr(ErrCodeScriptNameRequired, "脚本ID与名称不能为空")
	}
	if !model.ValidScriptLangs()[sc.Language] {
		return newScriptErr(ErrCodeScriptLangInvalid, fmt.Sprintf("非法的脚本语言: %s", sc.Language))
	}
	var exist model.JobScript
	if err := global.GVA_DB.First(&exist, sc.ID).Error; err != nil {
		return newScriptErr(ErrCodeScriptNotFound, "脚本不存在")
	}
	var count int64
	global.GVA_DB.Model(&model.JobScript{}).Where("name = ? AND id <> ?", sc.Name, sc.ID).Count(&count)
	if count > 0 {
		return newScriptErr(ErrCodeScriptDuplicate, fmt.Sprintf("脚本已存在: %s", sc.Name))
	}
	newVersion := exist.Version + 1
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.JobScript{}).Where("id = ?", sc.ID).Updates(map[string]any{
			"name": sc.Name, "language": sc.Language, "content": sc.Content,
			"version": newVersion, "notes": sc.Notes,
		}).Error; err != nil {
			return err
		}
		return tx.Create(&model.JobScriptVersion{ScriptID: sc.ID, Version: newVersion, Content: sc.Content, Operator: operator}).Error
	})
}

// DeleteScript 删除脚本（联动清版本表）
func (s *ScriptService) DeleteScript(id uint) error {
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("script_id = ?", id).Delete(&model.JobScriptVersion{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.JobScript{}, id).Error
	})
}

// GetScriptList 脚本列表（keyword 过滤，含 content 供执行页填充）
func (s *ScriptService) GetScriptList(keyword, language string) (list []*model.JobScript, err error) {
	db := global.GVA_DB.Model(&model.JobScript{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR notes LIKE ?", kw, kw)
	}
	if language != "" {
		db = db.Where("language = ?", language)
	}
	err = db.Order("id DESC").Find(&list).Error
	return list, err
}

// GetScriptVersions 脚本历史版本
func (s *ScriptService) GetScriptVersions(scriptID uint) (list []*model.JobScriptVersion, err error) {
	err = global.GVA_DB.Where("script_id = ?", scriptID).Order("version DESC").Limit(50).Find(&list).Error
	return list, err
}

// newScriptErr job 插件错误（避免与其他服务错误类型耦合）
func newScriptErr(code int, msg string) error { return &scriptError{Code: code, Msg: msg} }

type scriptError struct {
	Code int
	Msg  string
}

func (e *scriptError) Error() string { return e.Msg }

// VariableGroupService 变量组服务
type VariableGroupService struct{}

// validateVariableGroup 变量组校验（纯逻辑：名称必填、variables 为合法 JSON 数组）
func validateVariableGroup(vg *model.JobVariableGroup) error {
	if vg == nil || vg.Name == "" {
		return newScriptErr(ErrCodeVarGroupRequired, "变量组名称不能为空")
	}
	if vg.Variables != "" {
		var arr []map[string]string
		if err := json.Unmarshal([]byte(vg.Variables), &arr); err != nil {
			return newScriptErr(ErrCodeVariablesInvalid, "变量 JSON 格式不合法（应为 [{key,value}]）")
		}
	}
	return nil
}

// CreateVariableGroup 创建变量组（名称唯一）
func (s *VariableGroupService) CreateVariableGroup(vg *model.JobVariableGroup) error {
	if err := validateVariableGroup(vg); err != nil {
		return err
	}
	var count int64
	global.GVA_DB.Model(&model.JobVariableGroup{}).Where("name = ?", vg.Name).Count(&count)
	if count > 0 {
		return newScriptErr(ErrCodeVarGroupDuplicate, fmt.Sprintf("变量组已存在: %s", vg.Name))
	}
	return global.GVA_DB.Create(vg).Error
}

// DeleteVariableGroup 删除变量组
func (s *VariableGroupService) DeleteVariableGroup(id uint) error {
	return global.GVA_DB.Delete(&model.JobVariableGroup{}, id).Error
}

// UpdateVariableGroup 更新变量组
func (s *VariableGroupService) UpdateVariableGroup(vg *model.JobVariableGroup) error {
	if vg == nil || vg.ID == 0 || vg.Name == "" {
		return newScriptErr(ErrCodeVarGroupRequired, "变量组ID与名称不能为空")
	}
	if vg.Variables != "" {
		var arr []map[string]string
		if err := json.Unmarshal([]byte(vg.Variables), &arr); err != nil {
			return newScriptErr(ErrCodeVariablesInvalid, "变量 JSON 格式不合法")
		}
	}
	var count int64
	global.GVA_DB.Model(&model.JobVariableGroup{}).Where("name = ? AND id <> ?", vg.Name, vg.ID).Count(&count)
	if count > 0 {
		return newScriptErr(ErrCodeVarGroupDuplicate, fmt.Sprintf("变量组已存在: %s", vg.Name))
	}
	return global.GVA_DB.Model(&model.JobVariableGroup{}).Where("id = ?", vg.ID).Omit("ID").Updates(map[string]any{
		"name": vg.Name, "variables": vg.Variables, "asset_group_id": vg.AssetGroupID, "notes": vg.Notes,
	}).Error
}

// GetVariableGroupList 变量组全量列表
func (s *VariableGroupService) GetVariableGroupList(keyword string) (list []*model.JobVariableGroup, err error) {
	db := global.GVA_DB.Model(&model.JobVariableGroup{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR notes LIKE ?", kw, kw)
	}
	err = db.Order("id DESC").Find(&list).Error
	return list, err
}

// RenderTemplate 变量渲染：把 content 中 {{key}} 占位替换为变量组对应值（纯逻辑，可单测）
func RenderTemplate(content string, variables string) (string, error) {
	if variables == "" {
		return content, nil
	}
	var arr []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal([]byte(variables), &arr); err != nil {
		return content, newScriptErr(ErrCodeVariablesInvalid, "变量 JSON 格式不合法")
	}
	for _, kv := range arr {
		if kv.Key == "" {
			continue
		}
		content = strings.ReplaceAll(content, "{{"+kv.Key+"}}", kv.Value)
	}
	return content, nil
}
