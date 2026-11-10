// Package service ziqiang：侵权检测提示词库（需求 3.5-12/13）
// 口径：{{var}} 占位解析与渲染自实现（对齐 llmops G01/N12 口径），渲染缺失变量报错；
// 场景白名单=多模态双验证（文本相似度+图像特征）/引导式对话（创作场景→权利类型→法律文书）。
package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// 错误码（ziqiang 段 2700-2799）
const (
	ErrCodePromptParam   = 2716
	ErrCodePromptNotFound = 2717
)

// PromptScenes 场景白名单
var PromptScenes = []string{"多模态双验证", "引导式对话"}

func validPromptScene(s string) bool {
	for _, v := range PromptScenes {
		if v == s {
			return true
		}
	}
	return false
}

// varPattern {{var}} 占位
var varPattern = regexp.MustCompile(`\{\{(\w+)\}\}`)

// parsePromptVars 解析模板占位变量（纯函数，测试锚点）
func parsePromptVars(tpl string) []string {
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

// renderPrompt 模板渲染（纯函数，测试锚点）：缺失变量报错
func renderPrompt(tpl string, vars map[string]string) (string, error) {
	var missing []string
	for _, v := range parsePromptVars(tpl) {
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

// SavePromptTemplate 新建/更新模板（vars 自动解析，场景/步号白名单校验）
func (s *ZiqiangService) SavePromptTemplate(in *model.ZqPromptTemplate) (*model.ZqPromptTemplate, error) {
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Content) == "" {
		return nil, fmt.Errorf("[%d]模板名与内容必填", ErrCodePromptParam)
	}
	if !validPromptScene(in.Scene) {
		return nil, fmt.Errorf("[%d]场景须为 多模态双验证/引导式对话", ErrCodePromptParam)
	}
	if in.Step < 1 || in.Step > 3 {
		in.Step = 1
	}
	vars := parsePromptVars(in.Content)
	b, _ := json.Marshal(vars)
	in.Vars = string(b)
	if in.ID > 0 {
		var exist model.ZqPromptTemplate
		if err := global.GVA_DB.First(&exist, in.ID).Error; err != nil {
			return nil, fmt.Errorf("[%d]模板不存在", ErrCodePromptNotFound)
		}
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"name": in.Name, "scene": in.Scene, "step": in.Step,
			"content": in.Content, "vars": in.Vars, "enabled": in.Enabled,
		}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// DeletePromptTemplate 删除模板（软删除）
func (s *ZiqiangService) DeletePromptTemplate(id uint) error {
	res := global.GVA_DB.Delete(&model.ZqPromptTemplate{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("[%d]模板不存在", ErrCodePromptNotFound)
	}
	return nil
}

// GetPromptTemplates 模板列表（场景过滤，步号升序）
func (s *ZiqiangService) GetPromptTemplates(scene string) ([]model.ZqPromptTemplate, error) {
	db := global.GVA_DB.Model(&model.ZqPromptTemplate{})
	if scene != "" {
		if !validPromptScene(scene) {
			return nil, fmt.Errorf("[%d]场景须为 多模态双验证/引导式对话", ErrCodePromptParam)
		}
		db = db.Where("scene = ?", scene)
	}
	var list []model.ZqPromptTemplate
	err := db.Order("scene, step, id").Limit(500).Find(&list).Error
	return list, err
}

// RenderPromptByName 按模板名渲染（返回渲染结果与解析出的占位变量）
func (s *ZiqiangService) RenderPromptByName(name string, vars map[string]string) (map[string]any, error) {
	var tpl model.ZqPromptTemplate
	if err := global.GVA_DB.Where("name = ?", name).First(&tpl).Error; err != nil {
		return nil, fmt.Errorf("[%d]模板不存在: %s", ErrCodePromptNotFound, name)
	}
	rendered, err := renderPrompt(tpl.Content, vars)
	if err != nil {
		return nil, err
	}
	return map[string]any{"name": tpl.Name, "scene": tpl.Scene, "step": tpl.Step, "rendered": rendered, "vars": tpl.Vars}, nil
}
