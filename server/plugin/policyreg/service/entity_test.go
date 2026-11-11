// Package service policyreg：实体抽取单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
)

// TestExtractPolicyEntities 机构/文件号/日期
func TestExtractPolicyEntities(t *testing.T) {
	text := "根据教育部教〔2026〕15号文件和沪教委〔2025〕8号文要求，复旦大学与同济大学自2026年9月1日起实行学分互认，由学位评定委员会监督。"
	entities := extractPolicyEntities(text)
	byType := map[string][]string{}
	for _, e := range entities {
		byType[e.EType] = append(byType[e.EType], e.Name)
	}
	// 文件号
	if len(byType["文件号"]) != 2 {
		t.Fatalf("应抽取 2 个文件号: %+v", byType["文件号"])
	}
	if !strings.Contains(byType["文件号"][0], "教〔2026〕15号") {
		t.Fatalf("文件号不符: %+v", byType["文件号"])
	}
	// 日期
	if len(byType["日期"]) != 1 || byType["日期"][0] != "2026年9月1日" {
		t.Fatalf("日期抽取不符: %+v", byType["日期"])
	}
	// 机构（词典命中 教育部/学位评定委员会 + 规则 复旦大学/同济大学）
	joined := strings.Join(byType["机构"], "|")
	if !strings.Contains(joined, "教育部") || !strings.Contains(joined, "复旦大学") || !strings.Contains(joined, "学位评定委员会") {
		t.Fatalf("机构抽取不符: %s", joined)
	}
	// 去重
	text2 := "教育部 教育部 教育部"
	if got := len(extractPolicyEntities(text2)); got != 1 {
		t.Fatalf("应去重为 1: %d", got)
	}
	// 来源字段
	for _, e := range entities {
		if e.Source != "词典" && e.Source != "规则" {
			t.Fatalf("来源字段非法: %+v", e)
		}
	}
	_ = model.PolicyEntity{}
}
