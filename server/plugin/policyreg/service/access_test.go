// Package service policyreg：模型鉴权单测（纯函数，无 DB）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
)

// TestCanAccessRole 密级序+分类过滤
func TestCanAccessRole(t *testing.T) {
	grants := []model.PolicyPermGrant{
		{RoleName: "教师", DocCategory: "教学", MaxLevel: "内部", Allowed: true},
		{RoleName: "教师", DocCategory: "综合", MaxLevel: "公开", Allowed: true},
		{RoleName: "审计员", DocCategory: "综合", MaxLevel: "涉密", Allowed: true},
		{RoleName: "访客", DocCategory: "综合", MaxLevel: "涉密", Allowed: false},
	}
	// 教师访问 教学类内部制度 → 放行
	if !canAccessRole(grants, "教师", "教学", "内部") {
		t.Fatal("教师访问教学内部应放行")
	}
	// 教师访问 教学类涉密 → 拒绝（超 MaxLevel）
	if canAccessRole(grants, "教师", "教学", "涉密") {
		t.Fatal("教师访问涉密应拒绝")
	}
	// 教师访问 科研类内部 → 只能靠综合公开 → 拒绝
	if canAccessRole(grants, "教师", "科研", "内部") {
		t.Fatal("教师科研内部应拒绝")
	}
	// 审计员全类涉密 → 放行
	if !canAccessRole(grants, "审计员", "财务", "涉密") {
		t.Fatal("审计员应全类放行")
	}
	// 访客 Allowed=false → 拒绝
	if canAccessRole(grants, "访客", "教学", "公开") {
		t.Fatal("Allowed=false 应拒绝")
	}
	// 无 grant 角色 → 拒绝
	if canAccessRole(grants, "路人", "教学", "公开") {
		t.Fatal("无授权应拒绝")
	}
	// 未知密级 → 拒绝
	if canAccessRole(grants, "审计员", "教学", "绝密") {
		t.Fatal("未知密级应拒绝")
	}
}
