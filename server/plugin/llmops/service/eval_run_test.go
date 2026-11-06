// Package service llmops 模型评估单测（指标解析/判分三语义/批量评测假模型）
package service

import (
	"strings"
	"testing"
)

func TestParseEvalMetric(t *testing.T) {
	// 默认 contains
	m, err := ParseEvalMetric("")
	if err != nil || m.Type != "contains" {
		t.Fatalf("默认指标异常: %+v %v", m, err)
	}
	// keyword 必须带词表
	if _, err := ParseEvalMetric(`{"type":"keyword"}`); err == nil || !strings.Contains(err.Error(), "词表") {
		t.Fatalf("keyword 无词表应拒绝: %v", err)
	}
	// 非法类型
	if _, err := ParseEvalMetric(`{"type":"bertscore"}`); err == nil {
		t.Fatal("非法类型应拒绝")
	}
	// 阈值夹取
	m, _ = ParseEvalMetric(`{"type":"keyword","keywords":["a"],"threshold":5}`)
	if m.Threshold != 0.6 {
		t.Fatalf("阈值应夹取 0.6: %v", m.Threshold)
	}
}

func TestScoreAnswer(t *testing.T) {
	// exact：归一全等
	m := &EvalMetric{Type: "exact"}
	if ScoreAnswer(m, "白泽平台", "  白泽平台 ") != 1 {
		t.Fatal("exact 应归一命中")
	}
	if ScoreAnswer(m, "白泽平台", "白泽运维平台") != 0 {
		t.Fatal("exact 不应部分命中")
	}
	// contains
	m = &EvalMetric{Type: "contains"}
	if ScoreAnswer(m, "微调", "SFT 微调是监督微调") != 1 {
		t.Fatal("contains 应命中")
	}
	// keyword 命中率 + 阈值
	m = &EvalMetric{Type: "keyword", Keywords: []string{"SFT", "LoRA", "RLHF"}, Threshold: 0.6}
	if got := ScoreAnswer(m, "", "SFT 和 LoRA 都是微调方法"); got != 2.0/3.0 {
		t.Fatalf("keyword 命中率异常: %v", got)
	}
	if got := ScoreAnswer(m, "", "全都没提到"); got != 0 {
		t.Fatalf("零命中应为 0: %v", got)
	}
}

// TestEvalSamples 假模型注入的批量评测链路
func TestEvalSamples(t *testing.T) {
	samples := []EvalSample{
		{Question: "q1", Expected: "答案一"},
		{Question: "q2", Expected: "答案二"},
		{Question: "q3", Expected: "答案三"},
	}
	m := &EvalMetric{Type: "contains", Threshold: 1}
	fake := func(q string) (string, error) {
		switch q {
		case "q1":
			return "这是答案一", nil
		case "q2":
			return "完全跑题", nil
		default:
			return "", contextErr{}
		}
	}
	results, pass, err := evalSamples(fake, samples, m)
	if err == nil {
		t.Fatalf("模型失败应上抛: %v %+v", err, results)
	}
	_ = pass
	// 全成功场景
	fakeOk := func(q string) (string, error) {
		if q == "q2" {
			return "答案二的扩展", nil
		}
		return "包含答案一与答案三", nil
	}
	results, pass, err = evalSamples(fakeOk, samples, m)
	if err != nil {
		t.Fatalf("评测失败: %v", err)
	}
	if pass != 3 || len(results) != 3 || results[1].Idx != 2 {
		t.Fatalf("通过数/明细异常: pass=%d %+v", pass, results)
	}
}

type contextErr struct{}

func (contextErr) Error() string { return "model down" }
