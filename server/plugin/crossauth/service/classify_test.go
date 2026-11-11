// Package service crossauth：数据分级单测（纯函数，无 DB）
package service

import (
	"testing"
)

// TestClassifyData 分级取最高+分类多者胜
func TestClassifyData(t *testing.T) {
	// 敏感
	cat, lv, rule := classifyData("2026 级成绩单导出（含身份证号）")
	if lv != "敏感" || cat != "教学" || rule == "" {
		t.Fatalf("成绩单应敏感/教学: %s %s %s", cat, lv, rule)
	}
	// 内部
	_, lv, _ = classifyData("实验室项目预算表")
	if lv != "内部" {
		t.Fatalf("预算应内部: %s", lv)
	}
	// 公开默认
	cat, lv, _ = classifyData("校园开放日海报")
	if lv != "公开" || cat != "管理" {
		t.Fatalf("无关键词应公开: %s %s", cat, lv)
	}
	// 分类多者胜：教学 2 命中 vs 管理 1
	cat, _, _ = classifyData("课程考试安排 教学计划 财务报销")
	if cat != "教学" {
		t.Fatalf("教学应多者胜: %s", cat)
	}
}
