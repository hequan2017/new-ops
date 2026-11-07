// Package service llmops 模板渲染单测（替换语义/内联渲染链路）
package service

import (
	"strings"
	"testing"
)

func TestRenderPromptTemplate(t *testing.T) {
	// 全量替换 + 空白容忍
	out, err := RenderTemplate("把 {{ text }} 从 {{source_lang}} 翻到 {{ text }}", map[string]string{
		"text": "你好", "source_lang": "中文",
	})
	if err != nil {
		t.Fatalf("渲染失败: %v", err)
	}
	if out != "把 你好 从 中文 翻到 你好" {
		t.Fatalf("渲染结果异常: %q", out)
	}
	// 缺失报错带清单
	_, err = RenderTemplate("{{a}} {{b}}", map[string]string{"a": "1"})
	if err == nil || !strings.Contains(err.Error(), "缺少变量: b") {
		t.Fatalf("缺失应报清单: %v", err)
	}
	// 无变量透传
	out, _ = RenderTemplate("固定文案", nil)
	if out != "固定文案" {
		t.Fatalf("无变量应透传: %q", out)
	}
}

// TestRenderPrompt_Inline 内联渲染链路（不触 DB）
func TestRenderPrompt_Inline(t *testing.T) {
	svc := &LlmopsServiceStruct{}
	rendered, vars, err := svc.RenderPrompt(&RenderPromptIn{
		Content: "为 {{product}} 写文案，风格 {{tone}}",
		Vars:    map[string]string{"product": "白泽", "tone": "专业"},
	})
	if err != nil {
		t.Fatalf("内联渲染失败: %v", err)
	}
	if rendered != "为 白泽 写文案，风格 专业" {
		t.Fatalf("渲染异常: %q", rendered)
	}
	if len(vars) != 2 || vars[0] != "product" || vars[1] != "tone" {
		t.Fatalf("变量清单异常: %v", vars)
	}
	// 双空入参
	if _, _, err := svc.RenderPrompt(&RenderPromptIn{}); err == nil {
		t.Fatal("双空入参应拒绝")
	}
	// 缺失变量
	if _, _, err := svc.RenderPrompt(&RenderPromptIn{Content: "{{x}}"}); err == nil {
		t.Fatal("缺失变量应拒绝")
	}
}
