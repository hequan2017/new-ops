// Package service llmops 调试域单测（工具注册表快照隔离）
package service

import (
	"context"
	"testing"
)

// snapshotTool 快照验证用工具
type snapshotTool struct{ name string }

func (t *snapshotTool) Name() string                    { return t.name }
func (t *snapshotTool) Description() string             { return "snapshot test" }
func (t *snapshotTool) Execute(context.Context, string) (string, error) { return "", nil }

// TestCollectAgentTools 快照隔离：修改快照不影响注册表
func TestCollectAgentTools(t *testing.T) {
	RegisterAgentTool(&snapshotTool{name: "snap-a"})
	snap := collectAgentTools()
	if _, ok := snap["snap-a"]; !ok {
		t.Fatal("快照应含已注册工具")
	}
	delete(snap, "snap-a")
	snap["ghost"] = &snapshotTool{name: "ghost"}
	again := collectAgentTools()
	if _, ok := again["snap-a"]; !ok {
		t.Fatal("删除快照不应影响注册表")
	}
	if _, ok := again["ghost"]; ok {
		t.Fatal("快照污染不应写回注册表")
	}
}
