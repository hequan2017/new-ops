// Package service fanhua：提示词库（需求 3.4-18/19）
// 口径：{{var}} 占位解析与渲染自实现（对齐 llmops G01 口径）；
// 渲染时缺失变量报错；沪语场景模板与价值观审核链式三段模板（检测-修正-复核）种子幂等。
package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
	"github.com/hequan2017/new-ops/server/plugin/plugin-tool/errx"
)

// 错误码（fanhua 段 2600-2699）
const (
	ErrCodePromptParam    = 2623
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

// renderTemplate 模板渲染（纯函数，测试锚点）：缺失变量报错。
// 注：此处保留 "[码]" 前缀 fmt.Errorf 口径——既有单测锚定 err.Error() 含 "[2623]"，不可改测。
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
		return nil, errx.New(ErrCodePromptParam, "模板名与内容必填")
	}
	vars := parseTemplateVars(in.Content)
	b, _ := json.Marshal(vars)
	in.Vars = string(b)
	db := s.dbh()
	if in.ID > 0 {
		exist, err := s.repo.FindPromptTemplate(db, in.ID)
		if err != nil {
			return nil, errx.New(ErrCodePromptNotFound, "模板不存在")
		}
		if err := s.repo.UpdatePromptTemplateColumns(db, exist.ID, map[string]any{
			"name": in.Name, "scene": in.Scene, "content": in.Content, "vars": in.Vars, "enabled": in.Enabled,
		}); err != nil {
			return nil, errx.Wrap(ErrCodeFanhuaStore, "更新模板失败", err)
		}
		return exist, nil
	}
	if err := s.repo.CreatePromptTemplate(db, in); err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "创建模板失败", err)
	}
	return in, nil
}

// DeletePromptTemplate 删除模板（软删除）
func (s *FanhuaService) DeletePromptTemplate(id uint) error {
	rows, err := s.repo.DeletePromptTemplate(s.dbh(), id)
	if err != nil {
		return errx.Wrap(ErrCodeFanhuaStore, "删除模板失败", err)
	}
	if rows == 0 {
		return errx.New(ErrCodePromptNotFound, "模板不存在")
	}
	return nil
}

// GetPromptTemplates 模板列表（场景过滤）
func (s *FanhuaService) GetPromptTemplates(scene string) ([]model.FanhuaPromptTemplate, error) {
	list, err := s.repo.PromptTemplateList(s.dbh(), scene)
	if err != nil {
		return nil, errx.Wrap(ErrCodeFanhuaStore, "查询模板失败", err)
	}
	return list, nil
}

// RenderPromptByName 按模板名渲染
func (s *FanhuaService) RenderPromptByName(name string, vars map[string]string) (map[string]any, error) {
	tpl, err := s.repo.FindPromptTemplateByName(s.dbh(), name)
	if err != nil {
		return nil, errx.New(ErrCodePromptNotFound, "模板不存在: "+name)
	}
	rendered, err := renderTemplate(tpl.Content, vars)
	if err != nil {
		return nil, err
	}
	return map[string]any{"name": tpl.Name, "rendered": rendered, "vars": tpl.Vars}, nil
}
