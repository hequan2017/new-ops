// Package service llmops Prompt 优化单测（优化指令协议/批量逐条容错）
package service

import (
	"strings"
	"testing"
)

func TestBuildOptimizePrompt(t *testing.T) {
	// 有需求
	p := buildOptimizePrompt("原始内容", "更简洁，面向儿童")
	for _, kw := range []string{"更简洁，面向儿童", "原 Prompt：", "原始内容", "{{var}} 占位符原样"} {
		if !strings.Contains(p, kw) {
			t.Fatalf("优化指令缺 %q", kw)
		}
	}
	// 无需求 → 通用优化
	p2 := buildOptimizePrompt("x", "  ")
	if !strings.Contains(p2, "通用优化") {
		t.Fatalf("空需求应走通用优化: %q", p2)
	}
}

// TestOptimizeWith 批量逐条：失败留错不中断、空返回留错、原版保留对比
func TestOptimizeWith(t *testing.T) {
	items := []OptimizeItem{
		{Content: "第一条", Requirement: "改好一点"},
		{Content: "第二条"},                       // 模型返回空 → 留错
		{TemplateId: 999},                      // resolve 失败 → 留错
	}
	fn := func(prompt string) (string, error) {
		if strings.Contains(prompt, "第一条") {
			return "优化后的第一条", nil
		}
		return "  ", nil // 空返回
	}
	resolve := func(item *OptimizeItem) (string, error) {
		if item.TemplateId > 0 {
			return "", &testErr{"模板不存在"}
		}
		return item.Content, nil
	}
	results := optimizeWith(fn, items, resolve)
	if len(results) != 3 {
		t.Fatalf("应逐条产出 3 结果: %d", len(results))
	}
	if results[0].Err != "" || results[0].Original != "第一条" || results[0].Optimized != "优化后的第一条" {
		t.Fatalf("首条异常: %+v", results[0])
	}
	if results[1].Err == "" || results[1].Original != "第二条" {
		t.Fatalf("空返回应留错且保留原版: %+v", results[1])
	}
	if results[2].Err == "" || results[2].Original != "" {
		t.Fatalf("resolve 失败应留错: %+v", results[2])
	}
}

type testErr struct{ msg string }

func (e *testErr) Error() string { return e.msg }
