// Package service llmops 智能体运行时单测（JSON 截取/计划解析/单轮运行 fake 注入）
package service

import (
	"strings"
	"testing"
)

func TestExtractJSONObject(t *testing.T) {
	// 纯 JSON
	obj, ok := extractJSONObject(`{"intent":"a"}`)
	if !ok || obj != `{"intent":"a"}` {
		t.Fatalf("纯 JSON 异常: %q %v", obj, ok)
	}
	// 围栏 + 前后散文
	obj, ok = extractJSONObject("好的，如下：\n```json\n{\"intent\":\"查 询\",\"steps\":[\"x\"]}\n```\n完毕")
	if !ok || obj != `{"intent":"查 询","steps":["x"]}` {
		t.Fatalf("围栏截取异常: %q", obj)
	}
	// 嵌套花括号均衡（含字符串内的 }）
	obj, ok = extractJSONObject(`前缀 {"a":{"b":"}"},"c":1} 后缀`)
	if !ok || obj != `{"a":{"b":"}"},"c":1}` {
		t.Fatalf("嵌套截取异常: %q", obj)
	}
	// 无 JSON
	if _, ok := extractJSONObject("没有任何对象"); ok {
		t.Fatal("无对象应 false")
	}
}

func TestParseAgentPlan(t *testing.T) {
	// 直接解析
	p, err := ParseAgentPlan(`{"intent":"部署模型","steps":["选接入点","建容器"],"reply":"开始部署"}`)
	if err != nil || p.Intent != "部署模型" || len(p.Steps) != 2 {
		t.Fatalf("解析异常: %+v %v", p, err)
	}
	// 围栏兜底
	p, err = ParseAgentPlan("```json\n{\"intent\":\"查告警\"}\n```")
	if err != nil || p.Intent != "查告警" {
		t.Fatalf("围栏兜底异常: %+v %v", p, err)
	}
	// intent 缺失拒绝
	if _, err := ParseAgentPlan(`{"steps":[]}`); err == nil || !strings.Contains(err.Error(), "2270") {
		t.Fatalf("intent 缺失应报 2270: %v", err)
	}
	// 非 JSON
	if _, err := ParseAgentPlan("完全不是 JSON"); err == nil {
		t.Fatal("非 JSON 应拒绝")
	}
}

// TestRunAgentTurn qa 直答与拆解重试链路（fake 注入）
func TestRunAgentTurn(t *testing.T) {
	// qa：直答透传
	qaFn := func(_ string, msgs []ChatMessage) (string, error) {
		return "直答内容", nil
	}
	reply, plan, err := runAgentTurn(qaFn, "qa", "sys", nil, "问题", 5)
	if err != nil || reply != "直答内容" || plan != nil {
		t.Fatalf("qa 直答异常: %q %+v %v", reply, plan, err)
	}
	// agentic：首次坏输出 → 严格重试成功
	calls := 0
	retryFn := func(prompt string, _ []ChatMessage) (string, error) {
		calls++
		if calls == 1 {
			return "我不会输出 JSON", nil
		}
		if calls == 2 {
			if !strings.Contains(prompt, "禁止任何其它文字") {
				t.Fatalf("重试应带严格指令")
			}
			return `{"intent":"意图X","steps":["s1"],"reply":"答案X"}`, nil
		}
		return "最终回答", nil
	}
	reply, plan, err = runAgentTurn(retryFn, "agentic", "", nil, "做点事", 3)
	if err != nil || plan == nil || plan.Intent != "意图X" || reply != "最终回答" || calls != 3 {
		t.Fatalf("拆解重试链路异常: %q %+v calls=%d err=%v", reply, plan, calls, err)
	}
	// 两次都失败 → 报 2270
	failFn := func(_ string, _ []ChatMessage) (string, error) { return "no json", nil }
	if _, _, err := runAgentTurn(failFn, "agentic", "", nil, "x", 1); err == nil || !strings.Contains(err.Error(), "2270") {
		t.Fatalf("双次失败应报 2270: %v", err)
	}
}
