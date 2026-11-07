// Package service llmops 数据集联动单测（用途映射/匹配/守卫）
package service

import (
	"strings"
	"testing"
)

func TestPurposeForJob(t *testing.T) {
	cases := map[string][]string{
		"sft_full":  {"sft"},
		"sft_lora":  {"sft"},
		"reward":    {"rlhf"},
		"rl":        {"rlhf"},
		"eval":      {"eval"},
		"compress":  nil,
		"undefined": nil,
	}
	for job, want := range cases {
		got := purposeForJob(job)
		if len(got) != len(want) {
			t.Fatalf("purposeForJob(%q)=%v want %v", job, got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("purposeForJob(%q)=%v want %v", job, got, want)
			}
		}
	}
}

func TestPurposeAllowed(t *testing.T) {
	// 空约束放行
	if !purposeAllowed("anything", nil) {
		t.Fatal("空约束应放行")
	}
	if !purposeAllowed("sft", []string{"sft", "rlhf"}) {
		t.Fatal("命中应放行")
	}
	if purposeAllowed("eval", []string{"sft"}) {
		t.Fatal("不匹配应拒绝")
	}
}

// TestValidateDatasetForUse_Guard 空数据集 ID 守卫（不触 DB）
func TestValidateDatasetForUse_Guard(t *testing.T) {
	svc := &LlmopsServiceStruct{}
	if _, err := svc.ValidateDatasetForUse(0, []string{"sft"}); err == nil || !strings.Contains(err.Error(), "2299") {
		t.Fatalf("空 ID 应报 2299: %v", err)
	}
}
