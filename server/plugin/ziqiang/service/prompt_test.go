// Package service ziqiang：提示词库单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestParsePromptVars 占位解析：去重保序
func TestParsePromptVars(t *testing.T) {
	vars := parsePromptVars("分析 {{a}} 与 {{b}}，再回看 {{a}}；无占位则空")
	if len(vars) != 2 || vars[0] != "a" || vars[1] != "b" {
		t.Fatalf("应解析出去重后的 a,b: %v", vars)
	}
	if got := parsePromptVars("没有占位符"); len(got) != 0 {
		t.Fatalf("无占位应为空: %v", got)
	}
}

// TestRenderPrompt 渲染成功/缺失变量报错
func TestRenderPrompt(t *testing.T) {
	out, err := renderPrompt("文本相似度 {{score}}，阈值 {{threshold}}。", map[string]string{"score": "0.86", "threshold": "0.80"})
	if err != nil || out != "文本相似度 0.86，阈值 0.80。" {
		t.Fatalf("渲染不符: %s, %v", out, err)
	}
	if _, err := renderPrompt("汉明距离 {{d}}", map[string]string{}); err == nil || !strings.Contains(err.Error(), "缺失占位变量") {
		t.Fatalf("缺失变量应报错: %v", err)
	}
}

// TestValidPromptScene 场景与步号白名单
func TestValidPromptScene(t *testing.T) {
	for _, s := range PromptScenes {
		if !validPromptScene(s) {
			t.Fatalf("%s 应在白名单", s)
		}
	}
	if validPromptScene("通用") {
		t.Fatal("白名单外应拒绝")
	}
}
