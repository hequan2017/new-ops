// Package service llmops 工作流引擎单测（定义校验/模板/条件六式/执行链路）
package service

import (
	"context"
	"strings"
	"testing"
)

func TestParseWfDef(t *testing.T) {
	// 合法定义
	def, err := ParseWfDef(`{"nodes":[{"id":"a","type":"llm","prompt":"处理 {input}"},{"id":"b","type":"tool","tool":"echo","args":"{input}"}]}`)
	if err != nil || len(def.Nodes) != 2 {
		t.Fatalf("合法定义异常: %v", err)
	}
	// 空/重复 ID
	if _, err := ParseWfDef(`{"nodes":[{"id":"a","type":"human"},{"id":"a","type":"human"}]}`); err == nil {
		t.Fatal("重复 ID 应拒绝")
	}
	// llm 缺 prompt
	if _, err := ParseWfDef(`{"nodes":[{"id":"a","type":"llm"}]}`); err == nil {
		t.Fatal("llm 缺 prompt 应拒绝")
	}
	// 未知类型
	if _, err := ParseWfDef(`{"nodes":[{"id":"a","type":"magic"}]}`); err == nil {
		t.Fatal("未知类型应拒绝")
	}
	if _, err := ParseWfDef(`{"nodes":[]}`); err == nil {
		t.Fatal("空节点应拒绝")
	}
}

func TestRenderTemplate(t *testing.T) {
	if got := renderTemplate("处理：{input}", "文本X"); got != "处理：文本X" {
		t.Fatalf("渲染异常: %q", got)
	}
	if got := renderTemplate("无占位", "x"); got != "无占位" {
		t.Fatalf("无占位应原样: %q", got)
	}
}

func TestEvalCondition(t *testing.T) {
	cases := []struct {
		expr, input string
		want        bool
	}{
		{"contains:告警", "发现告警信息", true},
		{"contains:告警", "一切正常", false},
		{"not_contains:告警", "一切正常", true},
		{"len>3", "abcd", true},
		{"len>3", "abc", false},
		{"len<5", "abc", true},
		{"empty", "  ", true},
		{"not_empty", "x", true},
	}
	for _, c := range cases {
		got, err := evalCondition(c.expr, c.input)
		if err != nil || got != c.want {
			t.Fatalf("evalCondition(%q,%q)=%v,%v want %v", c.expr, c.input, got, err, c.want)
		}
	}
	if _, err := evalCondition("len>x", "abc"); err == nil {
		t.Fatal("非法数值应拒绝")
	}
	if _, err := evalCondition("regex:.*", "x"); err == nil {
		t.Fatal("不支持表达式应拒绝")
	}
}

// wfEcho 测试工具
type wfEcho struct{}

func (wfEcho) Name() string             { return "wfecho" }
func (wfEcho) Description() string      { return "echo" }
func (wfEcho) Execute(_ context.Context, argsJSON string) (string, error) {
	return "T(" + argsJSON + ")", nil
}

// TestRunWf_Linear llm→tool 线性链
func TestRunWf_Linear(t *testing.T) {
	def, _ := ParseWfDef(`{"nodes":[
		{"id":"a","type":"llm","prompt":"总结 {input}"},
		{"id":"b","type":"tool","tool":"wfecho","args":"{input}"}]}`)
	var gotPrompt string
	fn := func(prompt string, _ []ChatMessage) (string, error) {
		gotPrompt = prompt
		return "LLM 输出", nil
	}
	res, err := RunWf(fn, map[string]AgentTool{"wfecho": wfEcho{}}, def, "原始输入")
	if err != nil || res.Status != "done" || res.Output != "T(LLM 输出)" || len(res.Steps) != 2 {
		t.Fatalf("线性链异常: %+v %v", res, err)
	}
	if gotPrompt != "总结 原始输入" {
		t.Fatalf("prompt 渲染异常: %q", gotPrompt)
	}
}

// TestRunWf_ConditionJump 条件假→跳转路径（跳过中间节点）
func TestRunWf_ConditionJump(t *testing.T) {
	def, _ := ParseWfDef(`{"nodes":[
		{"id":"c","type":"condition","expr":"contains:紧急","onTrue":"llm1","onFalse":"end"},
		{"id":"llm1","type":"llm","prompt":"升级 {input}"},
		{"id":"end","type":"llm","prompt":"普通处理 {input}"}]}`)
	// 假 → 跳 end
	res, err := RunWf(func(string, []ChatMessage) (string, error) { return "ok", nil }, map[string]AgentTool{}, def, "普通事件")
	if err != nil || res.Status != "done" || res.Output != "ok" || len(res.Steps) != 2 {
		t.Fatalf("条件假路径异常: %+v %v", res, err)
	}
	if res.Steps[1].NodeId != "end" {
		t.Fatalf("应跳到 end: %+v", res.Steps)
	}
	// 真 → 走 llm1
	res2, _ := RunWf(func(string, []ChatMessage) (string, error) { return "升级完成", nil }, map[string]AgentTool{}, def, "紧急事件")
	if res2.Steps[1].NodeId != "llm1" {
		t.Fatalf("条件真应走 llm1: %+v", res2.Steps)
	}
}

// TestRunWf_Human 人工节点挂起
func TestRunWf_Human(t *testing.T) {
	def, _ := ParseWfDef(`{"nodes":[{"id":"h","type":"human"},{"id":"b","type":"llm","prompt":"不该执行"}]}`)
	called := false
	res, err := RunWf(func(string, []ChatMessage) (string, error) { called = true; return "", nil }, map[string]AgentTool{}, def, "待审内容")
	if err != nil || res.Status != "needs_human" || !strings.Contains(res.Output, "[需人工确认]") {
		t.Fatalf("人工挂起异常: %+v %v", res, err)
	}
	if called {
		t.Fatal("human 后不应继续执行")
	}
}

// TestRunWf_CycleGuard 跳转环防护
func TestRunWf_CycleGuard(t *testing.T) {
	def, _ := ParseWfDef(`{"nodes":[
		{"id":"c1","type":"condition","expr":"not_empty","onTrue":"c2"},
		{"id":"c2","type":"condition","expr":"not_empty","onTrue":"c1"}]}`)
	_, err := RunWf(func(string, []ChatMessage) (string, error) { return "", nil }, map[string]AgentTool{}, def, "x")
	if err == nil || !strings.Contains(err.Error(), "成环") {
		t.Fatalf("跳转环应被防护: %v", err)
	}
}
