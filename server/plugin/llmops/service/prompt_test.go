// Package service llmops Prompt 模板单测（{{var}} 抽取/校验）
package service

import "testing"

func TestExtractTemplateVars(t *testing.T) {
	// 去重保序 + 空白容忍
	got := ExtractTemplateVars("把 {{ text }} 从 {{source_lang}} 翻成 {{ text }}，作者 {{_author}}")
	want := []string{"text", "source_lang", "_author"}
	if len(got) != len(want) {
		t.Fatalf("变量数异常: %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("顺序异常: %v", got)
		}
	}
	// 无变量
	if len(ExtractTemplateVars("纯文本，没有占位符")) != 0 {
		t.Fatal("无变量应为空")
	}
	// 非法名（数字开头/单括号/中文）不识别
	if got := ExtractTemplateVars("{{1bad}} {single} {{中文}}"); len(got) != 0 {
		t.Fatalf("非法名应忽略: %v", got)
	}
}

func TestValidateTemplateVars(t *testing.T) {
	content := "翻译 {{text}} 为 {{lang}}"
	if err := ValidateTemplateVars(content, map[string]string{"text": "x", "lang": "y"}); err != nil {
		t.Fatalf("齐备应放行: %v", err)
	}
	err := ValidateTemplateVars(content, map[string]string{"text": "x"})
	if err == nil || err.Error() != "[2296]缺少变量: lang" {
		t.Fatalf("缺失应列清单: %v", err)
	}
	// 无变量模板任意 vars 放行
	if err := ValidateTemplateVars("固定文案", map[string]string{}); err != nil {
		t.Fatalf("无变量应放行: %v", err)
	}
}
