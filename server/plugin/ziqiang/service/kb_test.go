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
	// 「保护期」bigram 同命中著作权法，专利法应凭更高分居首
	hits2 := rankRegulations(list, "专利保护期限")
	if len(hits2) != 2 || hits2[0].Title != "中华人民共和国专利法" || hits2[0].Score <= hits2[1].Score {
		t.Fatalf("专利法应以更高分居首: %+v", hits2)
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

// TestRankCases 判例 bigram 检索
func TestRankCases(t *testing.T) {
	list := []model.ZqCase{
		{Jurisdiction: "中国", Name: "AI 文生图著作权第一案", Summary: "Stable Diffusion 生成图片", Ruling: "用户智力投入，享有著作权"},
		{Jurisdiction: "美国", Name: "Thaler v. Perlmutter", Summary: "AI 发明人与作者身份", Ruling: "人类作者为要件"},
	}
	// 「AI」bigram 同命中 Thaler 案，文生图案应凭更高分居首（对齐 TestRankRegulations 口径）
	hits := rankCases(list, "AI 生成图片著作权")
	if len(hits) != 2 || hits[0].Name != "AI 文生图著作权第一案" || hits[0].Score <= hits[1].Score {
		t.Fatalf("文生图案应以更高分居首: %+v", hits)
	}
	if hits2 := rankCases(list, "量子通信"); len(hits2) != 0 {
		t.Fatalf("无关查询应零命中: %+v", hits2)
	}
}
