// Package service llmops 训练任务单测（状态机流转/类型白名单/超参校验）
package service

import "testing"

// TestCanTransition 状态机契约表
func TestCanTransition(t *testing.T) {
	legal := []struct{ from, to string }{
		{"draft", "queued"}, {"draft", "failed"},
		{"queued", "running"}, {"queued", "failed"},
		{"running", "succeeded"}, {"running", "failed"}, {"running", "stopped"},
		{"stopped", "queued"}, {"failed", "queued"},
	}
	for _, c := range legal {
		if !CanTransition(c.from, c.to) {
			t.Fatalf("合法流转 %s→%s 被拒绝", c.from, c.to)
		}
	}
	illegal := []struct{ from, to string }{
		{"succeeded", "running"}, {"succeeded", "queued"},
		{"draft", "running"}, {"draft", "succeeded"},
		{"running", "draft"}, {"queued", "stopped"}, {"stopped", "running"},
	}
	for _, c := range illegal {
		if CanTransition(c.from, c.to) {
			t.Fatalf("非法流转 %s→%s 被放行", c.from, c.to)
		}
	}
}

func TestValidTrainType(t *testing.T) {
	for _, ok := range TrainTypes {
		if !ValidTrainType(ok) {
			t.Fatalf("合法类型 %q 被拒绝", ok)
		}
	}
	for _, bad := range []string{"", "SFT", "finetune", "sft_full;drop"} {
		if ValidTrainType(bad) {
			t.Fatalf("非法类型 %q 被放行", bad)
		}
	}
}

func TestValidateHyperJSON(t *testing.T) {
	if err := ValidateHyperJSON(""); err != nil {
		t.Fatalf("空超参应放行: %v", err)
	}
	if err := ValidateHyperJSON(`{"lr":0.001,"epochs":3}`); err != nil {
		t.Fatalf("合法对象应放行: %v", err)
	}
	if err := ValidateHyperJSON(`[1,2]`); err == nil {
		t.Fatal("数组应拒绝(须对象)")
	}
	if err := ValidateHyperJSON(`{bad`); err == nil {
		t.Fatal("坏 JSON 应拒绝")
	}
}
