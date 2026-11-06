// Package service llmops 智能体单测（模式/阶段白名单、轮次夹取、配置校验）
package service

import (
	"strings"
	"testing"
)

func TestValidAgentMode(t *testing.T) {
	for _, ok := range AgentModes {
		if !ValidAgentMode(ok) {
			t.Fatalf("合法模式 %q 被拒绝", ok)
		}
	}
	for _, bad := range []string{"", "QA", "agent", "workflow,x"} {
		if ValidAgentMode(bad) {
			t.Fatalf("非法模式 %q 被放行", bad)
		}
	}
}

func TestValidAgentStage(t *testing.T) {
	for _, ok := range AgentStages {
		if !ValidAgentStage(ok) {
			t.Fatalf("合法阶段 %q 被拒绝", ok)
		}
	}
	for _, bad := range []string{"", "publish", "DRAFT"} {
		if ValidAgentStage(bad) {
			t.Fatalf("非法阶段 %q 被放行", bad)
		}
	}
}

func TestClampMaxTurns(t *testing.T) {
	cases := map[int]int{0: 5, -3: 5, 1: 1, 5: 5, 20: 20, 21: 20, 100: 20}
	for in, want := range cases {
		if got := clampMaxTurns(in); got != want {
			t.Fatalf("clampMaxTurns(%d)=%d want %d", in, got, want)
		}
	}
}

// TestAgentConfigValidation 配置 JSON 对象契约（数组/标量拒绝，校验先于 DB 触达）
func TestAgentConfigValidation(t *testing.T) {
	svc := &LlmopsServiceStruct{}
	for _, bad := range []string{"[1,2]", `"str"`, "42", "{bad"} {
		var panicked bool
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			_, err := svc.SaveAgent(&AgentIn{Name: "x", Mode: "qa", ConfigJSON: bad})
			if panicked {
				t.Fatalf("非法配置应先报错而非触 DB: %q", bad)
			}
			if err == nil || !strings.Contains(err.Error(), "2267") {
				t.Fatalf("非法配置应报 2267: %q %v", bad, err)
			}
		}()
	}
}
