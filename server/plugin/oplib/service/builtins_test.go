// Package service oplib 内置算子单测（八算子契约）
package service

import (
	"strings"
	"testing"
)

func TestOpTextClean(t *testing.T) {
	out, err := opTextClean(map[string]any{"text": "  你好\t 世界 \x01混合\n\n空格  "})
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	got := out.(string)
	if strings.ContainsRune(got, '\x01') || strings.Contains(got, "\t") || strings.Contains(got, "  ") {
		t.Fatalf("清洗不彻底: %q", got)
	}
	if !strings.HasPrefix(got, "你好 世界") {
		t.Fatalf("首行异常: %q", got)
	}
}

func TestOpSplitSentences(t *testing.T) {
	out, _ := opSplitSentences(map[string]any{"text": "第一句。第二句！Third?第四；\n第五句"})
	got := out.([]string)
	if len(got) != 5 {
		t.Fatalf("应 5 句: %v", got)
	}
}

func TestOpDedupeLines(t *testing.T) {
	out, _ := opDedupeLines(map[string]any{"text": "a\nb\na\n\nb \n c"})
	got := out.(string)
	// 键归一去重（"b "与"b"同键），输出保留原行（" c" 首现原样）
	if got != "a\nb\n c" {
		t.Fatalf("去重异常: %q", got)
	}
}

func TestOpKeywordExtract(t *testing.T) {
	out, _ := opKeywordExtract(map[string]any{"text": "白泽平台 白泽平台 白泽", "topN": 3.0})
	got := out.([]map[string]any)
	if len(got) != 3 {
		t.Fatalf("应 3 个: %v", got)
	}
	if got[0]["term"] != "白泽" || got[0]["count"] != 3 {
		t.Fatalf("top1 异常: %v", got[0])
	}
}

func TestOpCountStats(t *testing.T) {
	out, _ := opCountStats(map[string]any{"text": "第一句。第二句！\n第三行"})
	got := out.(map[string]any)
	if got["sentences"] != 3 || got["lines"] != 2 {
		t.Fatalf("统计异常: %v", got)
	}
}

func TestOpJSON2CSV_and_Back(t *testing.T) {
	csvOut, err := opJSON2CSV(map[string]any{"data": []any{
		map[string]any{"name": "张三", "age": 20},
		map[string]any{"name": "李四", "age": 30},
	}})
	if err != nil {
		t.Fatalf("json2csv 失败: %v", err)
	}
	csvText := csvOut.(string)
	if !strings.HasPrefix(csvText, "age,name") || !strings.Contains(csvText, "20,张三") {
		t.Fatalf("CSV 内容异常: %q", csvText)
	}
	// 回转
	jsonOut, err := opCSV2JSON(map[string]any{"text": csvText})
	if err != nil {
		t.Fatalf("csv2json 失败: %v", err)
	}
	rows := jsonOut.([]map[string]string)
	if len(rows) != 2 || rows[1]["name"] != "李四" || rows[1]["age"] != "30" {
		t.Fatalf("回转异常: %v", rows)
	}
	// 非法 data
	if _, err := opJSON2CSV(map[string]any{"data": "not-array"}); err == nil {
		t.Fatal("非数组应拒绝")
	}
}

func TestOpRegexExtract(t *testing.T) {
	out, err := opRegexExtract(map[string]any{"pattern": `\d+`, "text": "a1b22c333"})
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	got := out.([]string)
	if len(got) != 3 || got[2] != "333" {
		t.Fatalf("提取异常: %v", got)
	}
	if _, err := opRegexExtract(map[string]any{"pattern": "([bad", "text": "x"}); err == nil {
		t.Fatal("坏正则应拒绝")
	}
	if _, err := opRegexExtract(map[string]any{"text": "x"}); err == nil {
		t.Fatal("空 pattern 应拒绝")
	}
}

func TestOpCaseNormalize(t *testing.T) {
	out, _ := opLowercase(map[string]any{"text": "Hello World", "mode": "upper"})
	if out.(string) != "HELLO WORLD" {
		t.Fatalf("upper 异常: %v", out)
	}
	out, _ = opLowercase(map[string]any{"text": "HELLO"}) // 缺省 lower
	if out.(string) != "hello" {
		t.Fatalf("缺省 lower 异常: %v", out)
	}
	if _, err := opLowercase(map[string]any{"text": "x", "mode": "weird"}); err == nil {
		t.Fatal("非法 mode 应拒绝")
	}
}

// TestBuiltinRegistry 注册表完整性（9 个：8 需求 + case_normalize 并入口径）
func TestBuiltinRegistry(t *testing.T) {
	want := []string{"text_clean", "split_sentences", "dedupe_lines", "keyword_extract",
		"count_stats", "json2csv", "csv2json", "regex_extract", "case_normalize"}
	for _, name := range want {
		if _, ok := GetBuiltin(name); !ok {
			t.Fatalf("内置算子 %s 未注册", name)
		}
	}
}
