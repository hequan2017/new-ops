// Package service llmops：Prompt 模板渲染（要求(7) API 调用 Prompt 模板）
// 口径：按 templateId 或内联 content 二选一；缺失变量报错列清单；返回渲染结果与所用变量。
package service

import (
	"fmt"
	"strings"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

// RenderTemplate 渲染（纯函数，测试锚点）：先校验齐备再替换
func RenderTemplate(content string, vars map[string]string) (string, error) {
	if err := ValidateTemplateVars(content, vars); err != nil {
		return "", err
	}
	return reTemplateVar.ReplaceAllStringFunc(content, func(m string) string {
		name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(m, "{{"), "}}"))
		return vars[name]
	}), nil
}

// RenderPromptIn 渲染请求
type RenderPromptIn struct {
	TemplateId uint            `json:"templateId"` // 与 content 二选一，优先模板
	Content    string          `json:"content"`
	Vars       map[string]string `json:"vars"`
}

// RenderPrompt 模板渲染服务（返回渲染结果与所用变量清单）
func (s *LlmopsServiceStruct) RenderPrompt(in *RenderPromptIn) (string, []string, error) {
	content := strings.TrimSpace(in.Content)
	if in.TemplateId > 0 {
		var tpl model.LlmopsPromptTemplate
		if err := global.GVA_DB.First(&tpl, in.TemplateId).Error; err != nil {
			return "", nil, fmt.Errorf("[%d]模板不存在", ErrCodePromptNotFound)
		}
		content = tpl.Content
	}
	if content == "" {
		return "", nil, fmt.Errorf("[%d]templateId 与 content 至少提供一个", ErrCodePromptNotFound)
	}
	rendered, err := RenderTemplate(content, in.Vars)
	if err != nil {
		return "", nil, err
	}
	return rendered, ExtractTemplateVars(content), nil
}
