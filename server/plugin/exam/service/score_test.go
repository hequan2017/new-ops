// Package service exam 成绩与等级单测（等级判定/CSV 形状）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/exam/model"
)

func TestDetermineLevel(t *testing.T) {
	// 默认档线
	def := defaultLevelRules()
	if got := DetermineLevel(def, 95, 100); got != "优秀" {
		t.Fatalf("95%% 应优秀: %s", got)
	}
	if got := DetermineLevel(def, 80, 100); got != "良好" {
		t.Fatalf("80%% 应良好: %s", got)
	}
	if got := DetermineLevel(def, 60, 100); got != "及格" {
		t.Fatalf("60%% 应及格: %s", got)
	}
	if got := DetermineLevel(def, 59.9, 100); got != "不及格" {
		t.Fatalf("59.9%% 应不及格: %s", got)
	}
	// 边界恰等
	if got := DetermineLevel(def, 90, 100); got != "优秀" {
		t.Fatalf("恰 90%% 应取优秀(≥): %s", got)
	}
	// 自定义规则覆盖（考试级）
	custom := []model.ExamLevelRule{{Level: "Pass", MinRatio: 0.5}, {Level: "Master", MinRatio: 0.8}}
	if got := DetermineLevel(custom, 85, 100); got != "Master" {
		t.Fatalf("自定义高档: %s", got)
	}
	if got := DetermineLevel(custom, 55, 100); got != "Pass" {
		t.Fatalf("自定义低档: %s", got)
	}
	if got := DetermineLevel(custom, 40, 100); got != "不及格" {
		t.Fatalf("无匹配默认不及格: %s", got)
	}
	// 满分为零防御
	if got := DetermineLevel(def, 0, 0); got != "不及格" {
		t.Fatalf("满分为零应不及格: %s", got)
	}
}

func TestScoresToCSV(t *testing.T) {
	rows := []model.ExamScore{
		{CandidateId: 1, Total: 88.5, Full: 100, Level: "良好"},
		{CandidateId: 2, Total: 92, Full: 100, Level: "优秀"},
	}
	csv := scoresToCSV(rows, map[uint]string{1: "张三", 2: "李四"})
	lines := strings.Split(strings.TrimSpace(csv), "\n")
	if len(lines) != 3 {
		t.Fatalf("表头+两行: %v", lines)
	}
	if !strings.HasPrefix(csv, "考生,总分,满分,等级") {
		t.Fatalf("表头异常: %q", csv)
	}
	if !strings.Contains(csv, "张三,88.5,100.0,良好") {
		t.Fatalf("行内容异常: %q", csv)
	}
	if !strings.Contains(csv, "李四,92.0,100.0,优秀") {
		t.Fatalf("行内容异常2: %q", csv)
	}
	// 未知考生名兜底为空
	csv2 := scoresToCSV(rows[:1], map[uint]string{})
	if !strings.Contains(csv2, ",88.5,100.0,良好") {
		t.Fatalf("缺名应保留数值列: %q", csv2)
	}
}
