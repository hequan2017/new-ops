// Package service llmops：Prompt 工程（要求(7)）——模板管理 + 变量占位符解析
// 口径：占位符 {{var}}（名称须为字母/下划线开头，容忍空白）；版本自增快照（G02 渲染面）。
package service

import (
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// 错误码（llmops 段续用）
const (
	ErrCodePromptNotFound = 2294
	ErrCodePromptNameDup  = 2295
	ErrCodePromptVarMiss  = 2296
)

var reTemplateVar = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)

// ExtractTemplateVars 抽取模板变量（去重保序，纯函数，测试锚点）
func ExtractTemplateVars(content string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range reTemplateVar.FindAllStringSubmatch(content, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	return out
}

// ValidateTemplateVars 变量齐备校验（纯函数：缺失列清单）
func ValidateTemplateVars(content string, vars map[string]string) error {
	var missing []string
	for _, v := range ExtractTemplateVars(content) {
		if _, ok := vars[v]; !ok {
			missing = append(missing, v)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("[%d]缺少变量: %s", ErrCodePromptVarMiss, strings.Join(missing, ","))
	}
	return nil
}

// GetPrompts 模板列表
func (s *LlmopsServiceStruct) GetPrompts() ([]model.LlmopsPromptTemplate, error) {
	var list []model.LlmopsPromptTemplate
	err := global.GVA_DB.Order("id").Find(&list).Error
	return list, err
}

// SavePrompt 新建/更新模板（更新版本自增并快照旧内容）
func (s *LlmopsServiceStruct) SavePrompt(in *model.LlmopsPromptTemplate) (*model.LlmopsPromptTemplate, error) {
	if in == nil || strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Content) == "" {
		return nil, fmt.Errorf("[%d]模板名与内容必填", ErrCodePromptNotFound)
	}
	var count int64
	q := global.GVA_DB.Model(&model.LlmopsPromptTemplate{}).Where("name = ?", in.Name)
	if in.ID > 0 {
		q = q.Where("id <> ?", in.ID)
	}
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("[%d]模板名已存在", ErrCodePromptNameDup)
	}
	if in.ID > 0 {
		var tpl model.LlmopsPromptTemplate
		if err := global.GVA_DB.First(&tpl, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]模板不存在", ErrCodePromptNotFound)
		}
		newVer := tpl.Version + 1
		err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&tpl).Updates(map[string]any{
				"name": in.Name, "content": in.Content, "desc": in.Desc, "version": newVer,
			}).Error; err != nil {
				return err
			}
			return tx.Create(&model.LlmopsPromptVersion{TemplateId: tpl.ID, Version: tpl.Version, Content: tpl.Content}).Error
		})
		if err != nil {
			return nil, err
		}
		tpl.Content, tpl.Version = in.Content, newVer
		return &tpl, nil
	}
	in.Version = 1
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeletePrompt 删除模板（级联删版本快照）
func (s *LlmopsServiceStruct) DeletePrompt(id uint) error {
	var tpl model.LlmopsPromptTemplate
	if err := global.GVA_DB.First(&tpl, id).Error; err != nil {
		return fmt.Errorf("[%d]模板不存在", ErrCodePromptNotFound)
	}
	err := global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("template_id = ?", id).Delete(&model.LlmopsPromptVersion{}).Error; err != nil {
			return err
		}
		return tx.Delete(&tpl).Error
	})
	return err
}

// GetPromptVersions 版本快照列表
func (s *LlmopsServiceStruct) GetPromptVersions(templateId uint) ([]model.LlmopsPromptVersion, error) {
	var list []model.LlmopsPromptVersion
	err := global.GVA_DB.Where("template_id = ?", templateId).Order("version desc").Find(&list).Error
	return list, err
}
