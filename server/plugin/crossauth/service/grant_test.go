// Package service crossauth：权限分配单测（纯函数，无 DB）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// TestAutoMatchRole 自动匹配：部门精确 > 通配兜底
func TestAutoMatchRole(t *testing.T) {
	rules := []model.CaAutoGrantRule{
		{School: "上海大学", Dept: "计算机学院", RoleName: "跨校科研人员"},
		{School: "上海大学", Dept: "*", RoleName: "跨校教师(教学)"},
		{School: "同济大学", Dept: "*", RoleName: "跨校学生(实训)"},
	}
	if r := autoMatchRole(rules, "上海大学", "计算机学院"); r == nil || r.RoleName != "跨校科研人员" {
		t.Fatalf("精确部门应优先: %+v", r)
	}
	if r := autoMatchRole(rules, "上海大学", "外国语学院"); r == nil || r.RoleName != "跨校教师(教学)" {
		t.Fatalf("应通配兜底: %+v", r)
	}
	if r := autoMatchRole(rules, "复旦大学", "计算机学院"); r != nil {
		t.Fatalf("无该校规则应 nil: %+v", r)
	}
}

// TestGrantNextStage 双重审核状态机
func TestGrantNextStage(t *testing.T) {
	stages := map[string]string{"提交": "本校审核", "本校审核": "归属校审核", "归属校审核": "通过"}
	for from, to := range stages {
		got, err := grantNextStage(from)
		if err != nil || got != to {
			t.Fatalf("%s 应推进 %s: %s %v", from, to, got, err)
		}
	}
	for _, terminal := range []string{"通过", "驳回", "未知"} {
		if _, err := grantNextStage(terminal); err == nil {
			t.Fatalf("%s 应不可推进", terminal)
		}
	}
}
