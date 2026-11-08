// Package service ziqiang：知识库检索单测（纯函数，无 DB）
package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/ziqiang/model"
)

// TestRankRegulations bigram 命中降序
func TestRankRegulations(t *testing.T) {
	list := []model.ZqRegulation{
		{Title: "中华人民共和国著作权法", Content: "作品自创作完成即享有著作权，保护期作者终生及死后五十年。"},
		{Title: "中华人民共和国专利法", Content: "发明专利保护期二十年，保护范围以权利要求为准。"},
		{Title: "GDPR", Content: "Personal data processing lawful basis and automated decision rights."},
	}
	hits := rankRegulations(list, "著作权保护期是多久")
	if len(hits) != 2 {
		t.Fatalf("两法规均含「保护期」应都命中: %d", len(hits))
	}
	if hits[0].Title != "中华人民共和国著作权法" || hits[0].Score <= hits[1].Score {
		t.Fatalf("著作权法应以更高分居首: %+v", hits)
	}
	hits2 := rankRegulations(list, "专利保护期限")
	if len(hits2) != 1 || hits2[0].Title != "中华人民共和国专利法" {
		t.Fatalf("专利法应命中: %+v", hits2)
	}
	if hits3 := rankRegulations(list, "量子纠缠"); len(hits3) != 0 {
		t.Fatalf("无关查询应零命中: %+v", hits3)
	}
}

// TestValidJurisdiction 辖区白名单
func TestValidJurisdiction(t *testing.T) {
	for _, j := range Jurisdictions {
		if !validJurisdiction(j) {
			t.Fatalf("%s 应在白名单", j)
		}
	}
	if validJurisdiction("日本") {
		t.Fatal("白名单外应拒绝")
	}
}
