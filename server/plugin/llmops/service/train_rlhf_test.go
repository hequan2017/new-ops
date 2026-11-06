// Package service llmops RLHF 两阶段单测（类型依赖校验纯函数）
package service

import (
	"strings"
	"testing"
)

func TestValidateRlhfTypes(t *testing.T) {
	// rl：前置必须 reward+succeeded
	if err := ValidateRlhfTypes("rl", "", "", false); err == nil || !strings.Contains(err.Error(), "2258") {
		t.Fatalf("rl 无前置应报 2258: %v", err)
	}
	if err := ValidateRlhfTypes("rl", "sft_lora", "succeeded", true); err == nil || !strings.Contains(err.Error(), "reward 类型") {
		t.Fatalf("rl 前置类型错应拒绝: %v", err)
	}
	if err := ValidateRlhfTypes("rl", "reward", "running", true); err == nil || !strings.Contains(err.Error(), "未完成") {
		t.Fatalf("rl 前置未完成应拒绝: %v", err)
	}
	if err := ValidateRlhfTypes("rl", "reward", "succeeded", true); err != nil {
		t.Fatalf("合法链路应放行: %v", err)
	}
	// reward：前置可选，仅接受 SFT 类
	if err := ValidateRlhfTypes("reward", "", "", false); err != nil {
		t.Fatalf("reward 无前置应放行: %v", err)
	}
	if err := ValidateRlhfTypes("reward", "sft_lora", "succeeded", true); err != nil {
		t.Fatalf("reward 前置 SFT 应放行: %v", err)
	}
	if err := ValidateRlhfTypes("reward", "rl", "succeeded", true); err == nil {
		t.Fatal("reward 前置 rl 应拒绝")
	}
	// 其余类型不校验链路
	if err := ValidateRlhfTypes("sft_full", "", "", false); err != nil {
		t.Fatalf("sft 不校验链路: %v", err)
	}
}
