// Package service crossauth：互认规则单测（纯函数，无 DB）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// TestTrustVerdict 互认判定
func TestTrustVerdict(t *testing.T) {
	rules := []model.CaTrustRule{
		{HomeIdentity: "上海大学教师", TargetResource: "合作高校教学资源库", Allowed: true},
		{HomeIdentity: "上海大学学生", TargetResource: "合作高校算力平台", Allowed: false},
		{HomeIdentity: "上海大学科研人员", TargetResource: "合作高校算力平台", Allowed: true},
	}
	// 教师→教学资源库 放行
	allow, reason := trustVerdict(rules, "上海大学教师", "教学资源库")
	if !allow {
		t.Fatalf("教师访问教学资源库应放行: %s", reason)
	}
	// 学生→算力 拒绝
	allow, _ = trustVerdict(rules, "上海大学学生", "算力平台")
	if allow {
		t.Fatal("学生访问算力应拒绝")
	}
	// 科研人员→算力 放行（同资源多规则取并集）
	allow, _ = trustVerdict(rules, "上海大学科研人员", "算力平台")
	if !allow {
		t.Fatal("科研人员访问算力应放行")
	}
	// 无匹配
	allow, _ = trustVerdict(rules, "校外人员", "教学资源库")
	if allow {
		t.Fatal("无匹配规则应拒绝")
	}
	// 空参数
	allow, _ = trustVerdict(rules, "", "教学资源库")
	if allow {
		t.Fatal("空身份应拒绝")
	}
}
