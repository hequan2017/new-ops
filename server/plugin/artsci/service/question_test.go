// Package service artsci：出题与组卷单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/artsci/model"
)

// TestTemplateQuestion 四题型模板生成
func TestTemplateQuestion(t *testing.T) {
	for _, qt := range []string{"单选", "多选", "判断", "简答"} {
		stem, answer, options := templateQuestion("相变动力学", qt)
		if !strings.Contains(stem, "相变动力学") || answer == "" {
			t.Fatalf("%s 模板题干/答案不符: %s / %s", qt, stem, answer)
		}
		if (qt == "单选" || qt == "多选") && len(options) != 4 {
			t.Fatalf("%s 应有 4 选项", qt)
		}
		if (qt == "判断" || qt == "简答") && options != nil {
			t.Fatalf("%s 不应有选项", qt)
		}
	}
	if _, answer, _ := templateQuestion("相变动力学", "单选"); answer != "A" {
		t.Fatalf("单选默认答案应为 A: %s", answer)
	}
}

// TestValidatePaperRules 规则校验与分值合计
func TestValidatePaperRules(t *testing.T) {
	rules, total, err := validatePaperRules(`[{"qType":"单选","count":5,"scorePer":2},{"knowledgePoint":"相变","qType":"简答","count":2,"scorePer":10}]`)
	if err != nil || total != 30 || len(rules) != 2 {
		t.Fatalf("合法规则应通过：total=%v err=%v", total, err)
	}
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"坏JSON", `{bad`, "须为 JSON 数组"},
		{"空规则", `[]`, "不能为空"},
		{"题型白名单外", `[{"qType":"口试","count":1,"scorePer":2}]`, "参数非法"},
		{"题量越界", `[{"qType":"单选","count":0,"scorePer":2}]`, "参数非法"},
		{"分值非法", `[{"qType":"单选","count":1,"scorePer":0}]`, "参数非法"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := validatePaperRules(tc.raw)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("期望含 %q，实际: %v", tc.want, err)
			}
		})
	}
}

// TestPickQuestions 抽题截取
func TestPickQuestions(t *testing.T) {
	candidates := make([]model.ArtsciQuestion, 3)
	picked := pickQuestions(candidates, 5)
	if len(picked) != 3 {
		t.Fatalf("候选不足应全取: %d", len(picked))
	}
	candidates = make([]model.ArtsciQuestion, 10)
	if got := len(pickQuestions(candidates, 4)); got != 4 {
		t.Fatalf("应取前 4 题: %d", got)
	}
}
