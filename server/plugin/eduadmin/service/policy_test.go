// Package service eduadmin：六大知识库单测（纯函数，无 DB）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/eduadmin/model"
)

// TestRankPolicies bigram 命中降序
func TestRankPolicies(t *testing.T) {
	list := []model.EduPolicy{
		{Title: "指标库-教学质量", Content: "生师比、评教均值、毕业率与学位授予率目标区间。"},
		{Title: "指标库-科研产出", Content: "论文数、纵向项目经费、专利授权量与转化率目标。"},
		{Title: "GDPR Summary", Content: "Personal data protection rules."},
	}
	hits := rankPolicies(list, "专利授权量与转化率")
	if len(hits) != 1 || hits[0].Title != "指标库-科研产出" {
		t.Fatalf("应仅科研产出命中: %+v", hits)
	}
	hits2 := rankPolicies(list, "指标目标区间")
	if len(hits2) < 2 || hits2[0].Score < hits2[1].Score {
		t.Fatalf("多命中应降序: %+v", hits2)
	}
	if hits3 := rankPolicies(list, "量子纠缠"); len(hits3) != 0 {
		t.Fatalf("无关查询应零命中: %+v", hits3)
	}
}

// TestValidEduPolicyDomain 六域白名单
func TestValidEduPolicyDomain(t *testing.T) {
	for _, d := range EduPolicyDomains {
		if !ValidEduPolicyDomain(d) {
			t.Fatalf("%s 应在白名单", d)
		}
	}
	if ValidEduPolicyDomain("招生") {
		t.Fatal("白名单外应拒绝")
	}
}
