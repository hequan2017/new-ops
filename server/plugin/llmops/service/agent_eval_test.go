// Package service llmops 智能体评测单测（评判解析/词表/注入式评测三型）
package service

import (
	"strings"
	"testing"
)

func TestParseAgentKeywords(t *testing.T) {
	got := ParseAgentKeywords(" SFT , LoRA ,, SFT,RLHF ")
	if len(got) != 3 || got[0] != "SFT" || got[2] != "RLHF" {
		t.Fatalf("词表解析异常: %v", got)
	}
	if len(ParseAgentKeywords("")) != 0 {
		t.Fatal("空词表应为空")
	}
}

func TestParseJudgeOutput(t *testing.T) {
	score, reason, err := ParseJudgeOutput(`{"score":0.8,"reason":"基本正确"}`)
	if err != nil || score != 0.8 || reason != "基本正确" {
		t.Fatalf("评判解析异常: %v %q %v", score, reason, err)
	}
	// 围栏 + 越界夹取
	score, _, err = ParseJudgeOutput("```json\n{\"score\":1.7}\n```")
	if err != nil || score != 1 {
		t.Fatalf("越界应夹取 1: %v %v", score, err)
	}
	score, _, _ = ParseJudgeOutput(`{"score":-0.5}`)
	if score != 0 {
		t.Fatalf("负分应夹取 0: %v", score)
	}
	// 缺 score/坏 JSON
	if _, _, err := ParseJudgeOutput(`{"reason":"无分数"}`); err == nil {
		t.Fatal("缺 score 应拒绝")
	}
	if _, _, err := ParseJudgeOutput("no json"); err == nil {
		t.Fatal("非 JSON 应拒绝")
	}
}

// TestEvalAgentSamples 注入式三型评测
func TestEvalAgentSamples(t *testing.T) {
	samples := []EvalSample{
		{Question: "q1", Expected: "微调"},
		{Question: "q2", Expected: "部署"},
	}
	// contains
	deps := &agentEvalDeps{replyFn: func(q string) (string, error) {
		if q == "q1" {
			return "SFT 微调方法", nil
		}
		return "跑题回答", nil
	}}
	results, pass, err := evalAgentSamples(deps, samples, "contains", nil)
	if err != nil || pass != 1 || results[0].Pass != true || results[1].Pass != false {
		t.Fatalf("contains 评测异常: pass=%d %+v %v", pass, results, err)
	}
	// keyword 命中率
	kw := []string{"微调", "LoRA"}
	deps2 := &agentEvalDeps{replyFn: func(string) (string, error) { return "SFT 微调与 LoRA 微调", nil }}
	results2, pass2, _ := evalAgentSamples(deps2, samples, "keyword", kw)
	if results2[0].Score != 1 || results2[0].Detail != "词表命中 2/2" {
		t.Fatalf("keyword 评分异常: %+v", results2[0])
	}
	_ = pass2
	// llm 评判
	deps3 := &agentEvalDeps{
		replyFn: func(string) (string, error) { return "回答内容", nil },
		judgeFn: func(_, _, _ string) (float64, string, error) { return 0.7, "大体正确", nil },
	}
	results3, pass3, err := evalAgentSamples(deps3, samples, "llm", nil)
	if err != nil || pass3 != 2 || results3[0].Score != 0.7 || results3[0].Detail != "大体正确" {
		t.Fatalf("llm 评测异常: pass=%d %+v %v", pass3, results3[0], err)
	}
	// 被测失败上抛
	deps4 := &agentEvalDeps{replyFn: func(string) (string, error) { return "", contextErr{} }}
	if _, _, err := evalAgentSamples(deps4, samples, "contains", nil); err == nil || !strings.Contains(err.Error(), "被测失败") {
		t.Fatalf("被测失败应上抛: %v", err)
	}
}
