package service

import (
	"encoding/json"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/workflow/model"
)

// threeNode 三节点经典审批流：提交 → 审批(approve/reject) → 归档
func threeNode(t *testing.T) (states string, transitions []model.WfTransition) {
	states = `[{"name":"submitted"},{"name":"approving","isApproval":true},{"name":"archived","isFinal":true}]`
	if err := json.Unmarshal([]byte(
		`[{"from":"submitted","action":"submit","to":"approving"},
		  {"from":"approving","action":"approve","to":"archived"},
		  {"from":"approving","action":"reject","to":"archived"}]`), &transitions); err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return states, transitions
}

func TestValidateDefinition(t *testing.T) {
	states, transitions := threeNode(t)
	trJSON, _ := json.Marshal(transitions)
	if _, _, err := ValidateDefinition(states, string(trJSON), "submitted"); err != nil {
		t.Fatalf("合法定义不应报错: %v", err)
	}
	if _, _, err := ValidateDefinition(states, string(trJSON), "nope"); err == nil {
		t.Fatal("起始状态不在状态集应报错")
	}
	if _, _, err := ValidateDefinition(states, `[{"from":"submitted","action":"x","to":"ghost"}]`, "submitted"); err == nil {
		t.Fatal("迁移引用不存在状态应报错")
	}
	if _, _, err := ValidateDefinition(`[{"name":"a"},{"name":"a"}]`, `[]`, "a"); err == nil {
		t.Fatal("状态重名应报错")
	}
}

func TestNextState(t *testing.T) {
	_, transitions := threeNode(t)
	if got, ok := NextState(transitions, "approving", "approve"); !ok || got != "archived" {
		t.Fatalf("approve 应到 archived: %s %v", got, ok)
	}
	if got, ok := NextState(transitions, "submitted", "submit"); !ok || got != "approving" {
		t.Fatalf("submit 应到 approving: %s %v", got, ok)
	}
	if _, ok := NextState(transitions, "submitted", "approve"); ok {
		t.Fatal("submitted 不接受 approve")
	}
}
