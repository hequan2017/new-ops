// Package service exam 题库单测（题型白名单/JSONL 解析/校验）
package service

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidQuestionType(t *testing.T) {
	for _, ok := range QuestionTypes {
		if !ValidQuestionType(ok) {
			t.Fatalf("合法题型 %q 被拒绝", ok)
		}
	}
	for _, bad := range []string{"", "SINGLE", "fill", "single;"} {
		if ValidQuestionType(bad) {
			t.Fatalf("非法题型 %q 被放行", bad)
		}
	}
}

func TestParseQuestionJSONL(t *testing.T) {
	lines := strings.Join([]string{
		`{"type":"single","stem":"SFT 是什么","options":["监督微调","强化学习"],"answer":"监督微调","score":2,"knowledgePoint":"微调","difficulty":2}`,
		`{"type":"judge","stem":"LoRA 是全参微调","answer":"错"}`,
		`这不是 JSON`, // 坏行跳过
		`{"type":"practice","stem":"用智能体完成一次 Pod 诊断","answer":"过程留痕"}`,
		`{"type":"single","stem":"缺选项","options":[]}`, // 校验失败跳过
	}, "\n")
	qs, skipped, err := ParseQuestionJSONL(lines)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(qs) != 3 || skipped != 2 {
		t.Fatalf("qs=%d skipped=%d want 3/2", len(qs), skipped)
	}
	// 选择题选项序列化
	var opts []string
	if err := jsonUnmarshal(qs[0].OptionsJSON, &opts); err != nil || len(opts) != 2 {
		t.Fatalf("选项序列化异常: %s", qs[0].OptionsJSON)
	}
	// 非选择题缺选项应放行
	if qs[1].OptionsJSON != "" {
		t.Fatalf("判断题不应强制选项: %q", qs[1].OptionsJSON)
	}
	// 全坏
	if _, _, err := ParseQuestionJSONL("bad\nbad2"); err == nil {
		t.Fatal("全坏应报错")
	}
}

func jsonUnmarshal(s string, v any) error {
	return json.Unmarshal([]byte(s), v)
}
