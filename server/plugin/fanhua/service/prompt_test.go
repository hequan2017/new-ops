// Package service fanhua：提示词渲染单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestParseTemplateVars 占位解析（去重）
func TestParseTemplateVars(t *testing.T) {
	vars := parseTemplateVars("{{a}} 与 {{b}}，再来一次 {{a}}；无占位 {x}")
	if len(vars) != 2 || vars[0] != "a" || vars[1] != "b" {
		t.Fatalf("应去重解析出 a,b: %+v", vars)
	}
	if len(parseTemplateVars("没有占位符")) != 0 {
		t.Fatal("无占位应空")
	}
}

// TestRenderTemplate 渲染：填充/缺失报错/多占位
func TestRenderTemplate(t *testing.T) {
	out, err := renderTemplate("用上海话解释：{{text}}（来源：{{source}}）", map[string]string{"text": "白相", "source": "测试"})
	if err != nil || !strings.Contains(out, "白相") || !strings.Contains(out, "测试") {
		t.Fatalf("渲染不符: %s %v", out, err)
	}
	if _, err = renderTemplate("{{a}} 和 {{b}}", map[string]string{"a": "1"}); err == nil || !strings.Contains(err.Error(), "[2623]") {
		t.Fatalf("缺失变量应报 2623: %v", err)
	}
	// 多次出现同一占位全替换
	out2, _ := renderTemplate("{{x}}={{x}}", map[string]string{"x": "v"})
	if out2 != "v=v" {
		t.Fatalf("同占位应全替换: %s", out2)
	}
}
