// Package service llmops 多智能体单测（定义解析/串联流水线/并行汇聚容错）
package service

import (
	"strings"
	"testing"
)

func TestParseMultiDef(t *testing.T) {
	def, err := ParseMultiDef(`{"pattern":"sequential","subs":[{"agentId":2,"role":"分析","prompt":"分析 {input}"}]}`)
	if err != nil || def.Pattern != "sequential" || len(def.Subs) != 1 {
		t.Fatalf("合法定义异常: %+v %v", def, err)
	}
	if _, err := ParseMultiDef(`{"pattern":"debate","subs":[{"agentId":2}]}`); err == nil {
		t.Fatal("非法 pattern 应拒绝")
	}
	if _, err := ParseMultiDef(`{"pattern":"converge","subs":[]}`); err == nil {
		t.Fatal("空 subs 应拒绝")
	}
	many := `{"pattern":"converge","subs":[`
	for i := 0; i < 11; i++ {
		if i > 0 {
			many += ","
		}
		many += `{"agentId":2}`
	}
	many += `]}`
	if _, err := ParseMultiDef(many); err == nil {
		t.Fatal("超上限应拒绝")
	}
}

// TestRunMultiAgent_Sequential 串联流水线输出接力
func TestRunMultiAgent_Sequential(t *testing.T) {
	def := &MultiDef{Pattern: "sequential", Subs: []SubAgentRef{
		{AgentId: 2, Role: "提炼", Prompt: "提炼 {input}"},
		{AgentId: 3, Role: "审校", Prompt: "审校 {input}"},
	}}
	var seen []string
	run := func(sub SubAgentRef, input string) (string, error) {
		seen = append(seen, sub.Role+":"+input)
		if sub.Role == "提炼" {
			return "要点A", nil
		}
		return "终稿(" + input + ")", nil
	}
	res, err := runMultiAgent(def, run, nil, "原始稿")
	if err != nil || res.Output != "终稿(要点A)" || len(res.Subs) != 2 {
		t.Fatalf("串联链异常: %+v %v", res, err)
	}
	if seen[0] != "提炼:原始稿" || seen[1] != "审校:要点A" {
		t.Fatalf("接力顺序异常: %v", seen)
	}
}

// TestRunMultiAgent_Converge 并行汇聚：单子失败容忍 + 全败拒绝 + 汇聚输入断言
func TestRunMultiAgent_Converge(t *testing.T) {
	def := &MultiDef{Pattern: "converge", Subs: []SubAgentRef{
		{AgentId: 2, Role: "甲"}, {AgentId: 3, Role: "乙"}, {AgentId: 4, Role: "丙"},
	}}
	run := func(sub SubAgentRef, _ string) (string, error) {
		if sub.Role == "乙" {
			return "", contextErr{}
		}
		return "观点-" + sub.Role, nil
	}
	var aggInput string
	agg := func(prompt string, _ []ChatMessage) (string, error) {
		aggInput = prompt
		return "综合结论", nil
	}
	res, err := runMultiAgent(def, run, agg, "问题")
	if err != nil || res.Output != "综合结论" {
		t.Fatalf("汇聚异常: %+v %v", res, err)
	}
	if !strings.Contains(aggInput, "【甲】观点-甲") || !strings.Contains(aggInput, "【丙】观点-丙") || strings.Contains(aggInput, "观点-乙") {
		t.Fatalf("汇聚输入应含成功者排除失败者: %q", aggInput)
	}
	if res.Subs[1].Err == "" {
		t.Fatal("失败子应留痕")
	}
	// 全败
	runFail := func(SubAgentRef, string) (string, error) { return "", contextErr{} }
	if _, err := runMultiAgent(def, runFail, nil, "x"); err == nil || !strings.Contains(err.Error(), "全部子智能体") {
		t.Fatalf("全败应拒绝: %v", err)
	}
}
