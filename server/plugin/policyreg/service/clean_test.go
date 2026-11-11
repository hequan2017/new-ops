// Package service policyreg：制度清洗单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestCheckClauseConsistency 重复条号+冲突关键词对
func TestCheckClauseConsistency(t *testing.T) {
	text := "第一条 制度目的明确。第二条 应当遵守规定。第二条 重复出现。第三条 必须全程记录，禁止遗漏。\n第四条 正常条款。"
	issues := checkClauseConsistency(text)
	types := map[string]int{}
	for _, iss := range issues {
		types[iss.Type]++
	}
	if types["重复条号"] != 1 {
		t.Fatalf("应 1 处重复条号: %+v", issues)
	}
	if types["条款冲突"] != 1 {
		t.Fatalf("应 1 处条款冲突（必须×禁止）: %+v", issues)
	}
	// 干净文本
	if issues := checkClauseConsistency("第一条 规范管理。第二条 不得外传。"); len(issues) != 0 {
		t.Fatalf("干净文本应零问题: %+v", issues)
	}
}

// TestCheckFormatConflict 全半角句号混用
func TestCheckFormatConflict(t *testing.T) {
	if issues := checkFormatConflict("第一条 用中文句号。"); len(issues) != 0 {
		t.Fatalf("单一样式应无冲突: %+v", issues)
	}
	issues := checkFormatConflict("第一条 混用句号.Second version。")
	if len(issues) != 1 || !strings.Contains(issues[0].Detail, "混用") {
		t.Fatalf("应检出混用: %+v", issues)
	}
	// 小数点不误报
	if issues := checkFormatConflict("第一条 数值为 3.5，用中文句号。"); len(issues) != 0 {
		t.Fatalf("小数点不应误报: %+v", issues)
	}
}

// TestSplitClauses 条号切分
func TestSplitClauses(t *testing.T) {
	clauses := splitClauses("第一条 总则内容。第二条 分则内容。第十条 尾部。")
	if len(clauses) != 3 || clauses[0][0] != "一" || clauses[2][0] != "十" {
		t.Fatalf("切分不符: %+v", clauses)
	}
	if !strings.Contains(clauses[1][1], "分则内容") {
		t.Fatalf("条文本应含正文: %s", clauses[1][1])
	}
}
