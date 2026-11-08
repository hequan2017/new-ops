// Package service fanhua：提示词库（需求 3.4-18/19）
// 口径：{{var}} 占位解析与渲染自实现（对齐 llmops G01 口径）；
// 渲染时缺失变量报错；沪语场景模板与价值观审核链式三段模板（检测-修正-复核）种子幂等。
package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodePromptParam   = 2623
	ErrCodePromptNotFound = 2624
)

// varPattern {{var}} 占位
var varPattern = regexp.MustCompile(`\{\{(\w+)\}\}`)

// parseTemplateVars 解析模板占位变量（纯函数，测试锚点）
func parseTemplateVars(tpl string) []string {
	seen := map[string]bool{}
	var vars []string
	for _, m := range varPattern.FindAllStringSubmatch(tpl, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			vars = append(vars, m[1])
		}
	}
	return vars
}

// renderTemplate 模板渲染（纯函数，测试锚点）：缺失变量报错
func renderTemplate(tpl string, vars map[string]string) (string, error) {
	var missing []string
	for _, v := range parseTemplateVars(tpl) {
		if _, ok := vars[v]; !ok {
			missing = append(missing, v)
		}
	}
	if len(missing) > 0 {
		return "", fmt.Errorf("[%d]缺失占位变量: %s", ErrCodePromptParam, strings.Join(missing, ","))
	}
	return varPattern.ReplaceAllStringFunc(tpl, func(m string) string {
		name := varPattern.FindStringSubmatch(m)[1]
		return vars[name]
	}), nil
}

// SavePromptTemplate 新建/更新模板（vars 自动解析）
func (s *FanhuaService) SavePromptTemplate(in *model.FanhuaPromptTemplate) (*model.FanhuaPromptTemplate, error) {
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Content) == "" {
		return nil, fmt.Errorf("[%d]模板名与内容必填", ErrCodePromptParam)
	}
	vars := parseTemplateVars(in.Content)
	b, _ := json.Marshal(vars)
	in.Vars = string(b)
	if in.ID > 0 {
		var exist model.FanhuaPromptTemplate
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]模板不存在", ErrCodePromptNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"name": in.Name, "scene": in.Scene, "content": in.Content, "vars": in.Vars, "enabled": in.Enabled,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeletePromptTemplate 删除模板（软删除）
func (s *FanhuaService) DeletePromptTemplate(id uint) error {
	res := global.GVA_DB.Delete(&model.FanhuaPromptTemplate{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]模板不存在", ErrCodePromptNotFound)
	}
	return nil
}

// GetPromptTemplates 模板列表（场景过滤）
func (s *FanhuaService) GetPromptTemplates(scene string) ([]model.FanhuaPromptTemplate, error) {
	db := global.GVA_DB.Model(&model.FanhuaPromptTemplate{})
	if scene != "" {
		db = db.Where("scene = ?", scene)
	}
	var list []model.FanhuaPromptTemplate
	err := db.Order("id").Limit(500).Find(&list).Error
	return list, err
}

// RenderPromptByName 按模板名渲染
func (s *FanhuaService) RenderPromptByName(name string, vars map[string]string) (map[string]any, error) {
	var tpl model.FanhuaPromptTemplate
	if err := global.GVA_DB.Where("name = ?", name).First(&tpl).Error; err != nil {
		return nil, fmt.Errorf("[%d]模板不存在: %s", ErrCodePromptNotFound, name)
	}
	rendered, err := renderTemplate(tpl.Content, vars)
	if err != nil {
		return nil, err
	}
	return map[string]any{"name": tpl.Name, "rendered": rendered, "vars": tpl.Vars}, nil
}
