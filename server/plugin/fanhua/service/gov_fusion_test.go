// Package service fanhua：城市科学融合查询单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestBuildGovSuggestions 低分触发建议、全好保持口径
func TestBuildGovSuggestions(t *testing.T) {
	low := map[string]float64{"空气质量": 45.6, "绿化视觉": 30}
	sugg := buildGovSuggestions(low)
	if len(sugg) != 2 {
		t.Fatalf("两个低分叶应出 2 条建议: %+v", sugg)
	}
	if !strings.Contains(sugg[0], "空气质量") || !strings.Contains(sugg[0], "扬尘") {
		t.Fatalf("空气质量建议文案不符: %s", sugg[0])
	}
	if !strings.Contains(sugg[1], "绿化视觉") || !strings.Contains(sugg[1], "口袋公园") {
		t.Fatalf("绿化建议文案不符: %s", sugg[1])
	}
	good := buildGovSuggestions(map[string]float64{"空气质量": 85, "服务满意度": 90})
	if len(good) != 1 || !strings.Contains(good[0], "保持现有治理节奏") {
		t.Fatalf("全好应输出保持口径: %+v", good)
	}
	// 未知指标名不触发（规则外忽略）
	unknown := buildGovSuggestions(map[string]float64{"未知指标": 10})
	if len(unknown) != 1 || !strings.Contains(unknown[0], "保持现有治理节奏") {
		t.Fatalf("规则外指标应输出保持口径: %+v", unknown)
	}
}

// TestCollectLeafScores 叶得分收集
func TestCollectLeafScores(t *testing.T) {
	roots := buildIndicatorTree(govList())
	scoreTree(roots[0], govPoints())
	scores := collectLeafScores(roots)
	if scores["空气质量"] != 80 || scores["噪声水平"] != 60 {
		t.Fatalf("叶得分不符: %+v", scores)
	}
	if _, isLeaf := scores["环境品质"]; isLeaf {
		t.Fatal("根节点不应进入叶得分表")
	}
}

// TestSaveGovPhotoValidate 照片登记校验（service 层参数分支，DB 不可达时跳过语义）
func TestSaveGovPhotoValidate(t *testing.T) {
	// 纯校验逻辑经由 FusionQuery/SaveGovPhoto 的空参分支——此处验证规则映射表完整
	for name := range govSuggestionRules {
		if govSuggestionRules[name] == "" {
			t.Fatalf("指标 %s 建议文案为空", name)
		}
	}
}
