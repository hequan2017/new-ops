// Package service llmops 问答对采集纯函数单测（JSON/CSV 解析、格式探测、校验去重）
package service

import (
	"strings"
	"testing"
)

func TestDetectQAFormat(t *testing.T) {
	cases := map[string]string{
		`[{"question":"q","answer":"a"}]`: "json",
		"\n\n {\"x\":1}":                  "json",
		"\xef\xbb\xbf[1,2]":               "json", // BOM 前缀
		"question,answer\nq1,a1":          "csv",
		"":                                "csv",
	}
	for in, want := range cases {
		if got := DetectQAFormat([]byte(in)); got != want {
			t.Fatalf("DetectQAFormat(%q)=%q want %q", in, got, want)
		}
	}
}

func TestParseQaJSON(t *testing.T) {
	items, err := ParseQaJSON([]byte(`[{"question":"什么是白泽","answer":"统一运维平台"},{"q":"别名","a":"BaiZe"},{"question":"仅问题"}]`))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(items) != 3 || items[1].Question != "别名" || items[1].Answer != "BaiZe" {
		t.Fatalf("别名兼容失败: %+v", items)
	}
	if items[0].Line != 1 || items[2].Line != 3 {
		t.Fatalf("Line 下标错误: %+v", items)
	}
	if _, err := ParseQaJSON([]byte(`{"not":"array"}`)); err == nil || !strings.Contains(err.Error(), "2208") {
		t.Fatalf("非数组应报 2208: %v", err)
	}
}

func TestParseQaCSV(t *testing.T) {
	csvData := "question,answer\nQ1,A1\n\nq2,a2\ntoo_few\nq3,a3"
	items, skipped, err := ParseQaCSV([]byte(csvData))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(items) != 3 || skipped != 1 {
		t.Fatalf("items=%d skipped=%d want 3/1", len(items), skipped)
	}
	if items[0].Line != 2 || items[2].Line != 5 {
		t.Fatalf("CSV 行号错误: %+v", items)
	}
	// q/a 别名表头
	items2, _, err := ParseQaCSV([]byte("q,a\nx,y"))
	if err != nil || len(items2) != 1 {
		t.Fatalf("q/a 表头兼容失败: %v %+v", err, items2)
	}
	// 缺表头
	if _, _, err := ParseQaCSV([]byte("foo,bar\nx,y")); err == nil || !strings.Contains(err.Error(), "表头") {
		t.Fatalf("缺表头应报错: %v", err)
	}
}

func TestValidateDedupeQa(t *testing.T) {
	items := []QaItem{
		{Question: " 什么是  白泽？ ", Answer: "平台", Line: 1},
		{Question: "什么是 白泽？", Answer: "另一个答案", Line: 2}, // 归一后同题 → dup
		{Question: "WHAT IS  BAIZE", Answer: "ascii", Line: 3},
		{Question: "what is baize", Answer: "ASCII 小写同题", Line: 4}, // ToLower 同题 → dup
		{Question: "  ", Answer: "空问题", Line: 5},                    // 跳过
		{Question: "有答案吗", Answer: "", Line: 6},                    // 跳过
		{Question: "独立问题", Answer: "独立答案", Line: 7},
	}
	kept, skipped, dup, err := ValidateDedupeQa(items)
	if err != nil {
		t.Fatalf("校验失败: %v", err)
	}
	if len(kept) != 3 || skipped != 2 || dup != 2 {
		t.Fatalf("kept=%d skipped=%d dup=%d want 3/2/2", len(kept), skipped, dup)
	}
	if kept[0].Question != "什么是  白泽？" { // 首尾 trim 保留内部原空格
		t.Fatalf("问题应 trim 首尾: %q", kept[0].Question)
	}
	// 全部无效
	if _, _, _, err := ValidateDedupeQa([]QaItem{{Question: " ", Answer: " "}}); err == nil || !strings.Contains(err.Error(), "2209") {
		t.Fatalf("全无效应报 2209: %v", err)
	}
}

func TestNormalizeQaKey(t *testing.T) {
	cases := map[string]string{
		"A  B\tC": "a b c",
		"  X  ":   "x",
		"白泽 平台":   "白泽 平台",
	}
	for in, want := range cases {
		if got := NormalizeQaKey(in); got != want {
			t.Fatalf("NormalizeQaKey(%q)=%q want %q", in, got, want)
		}
	}
}
