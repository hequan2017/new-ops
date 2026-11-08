// Package service fanhua：文本标注单测（纯函数，无 DB）
package service

import (
	"testing"
)

// TestSentimentRule 情感三档
func TestSentimentRule(t *testing.T) {
	cases := []struct {
		text string
		want string
	}{
		{"迭个小囡开心得来，老灵额", "正面"},
		{"味道难吃，气死我了", "负面"},
		{"今朝天气一般般", "中立"},
		{"蛮满意但也交关失望", "中立"}, // 正负各 1 → 平局
	}
	for _, c := range cases {
		if got := sentimentRule(c.text); got != c.want {
			t.Fatalf("sentimentRule(%q)=%s 期望 %s", c.text, got, c.want)
		}
	}
}

// TestAssignSplit 确定性划分：同 id 同结果、比例近似
func TestAssignSplit(t *testing.T) {
	if assignSplit(42, 0.8) != assignSplit(42, 0.8) {
		t.Fatal("同 id 应确定性")
	}
	train := 0
	const n = 1000
	for id := uint(1); id <= n; id++ {
		if assignSplit(id, 0.8) == "train" {
			train++
		}
	}
	if train < 720 || train > 880 {
		t.Fatalf("0.8 比例下 train 应约 800: %d", train)
	}
	if assignSplit(7, 1.0) != "train" {
		t.Fatal("比例 1.0 应全 train")
	}
}

// TestValidateRatio 比例白名单
func TestValidateRatio(t *testing.T) {
	if err := validateRatio(0.8); err != nil {
		t.Fatalf("0.8 应合法: %v", err)
	}
	if err := validateRatio(0.95); err == nil {
		t.Fatal("0.95 应拒绝")
	}
}
