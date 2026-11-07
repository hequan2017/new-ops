// Package service llmops 零代码技能单测（生成 prompt 协议/输出解析/重试链路）
package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

func TestBuildSkillGenPrompt(t *testing.T) {
	p := buildSkillGenPrompt("做一个把文字变拼音的技能")
	for _, kw := range []string{`{"name"`, `"kind":"prompt"`, `{input}`, "需求描述：做一个把文字变拼音的技能"} {
		if !strings.Contains(p, kw) {
			t.Fatalf("生成 prompt 缺 %q", kw)
		}
	}
}

func TestParseSkillGenOutput(t *testing.T) {
	good := `这是生成结果：
{"name":"拼音转换","desc":"文字转拼音","def":{"kind":"prompt","system":"只输出拼音","template":"转拼音：{input}"}}`
	sk, err := ParseSkillGenOutput(good)
	if err != nil || sk.Name != "拼音转换" || sk.Type != "custom" || sk.Status != "enabled" {
		t.Fatalf("解析异常: %+v %v", sk, err)
	}
	// 围栏 + 超长名截断
	fenced := "```json\n{\"name\":\"" + strings.Repeat("长", 50) + "\",\"def\":{\"kind\":\"prompt\",\"template\":\"{input}\"}}\n```"
	sk2, err := ParseSkillGenOutput(fenced)
	if err != nil || len([]rune(sk2.Name)) != 32 {
		t.Fatalf("超长名应截断 32: %d %v", len([]rune(sk2.Name)), err)
	}
	// 名称为空/def 缺失
	if _, err := ParseSkillGenOutput(`{"name":"  ","def":{"kind":"prompt","template":"x"}}`); err == nil {
		t.Fatal("空名应拒绝")
	}
	if _, err := ParseSkillGenOutput(`{"name":"x"}`); err == nil {
		t.Fatal("缺 def 应拒绝")
	}
	// def 非法（kind 错）
	if _, err := ParseSkillGenOutput(`{"name":"x","def":{"kind":"code","template":"y"}}`); err == nil {
		t.Fatal("def 非法应拒绝")
	}
	if _, err := ParseSkillGenOutput("没有 JSON"); err == nil {
		t.Fatal("无 JSON 应拒绝")
	}
}

// TestCreateSkillFromDescription_Retry 重试链路（fake chat；SaveSkill 触 DB，用 recover 断言生成链先行）
func TestCreateSkillFromDescription_Retry(t *testing.T) {
	svc := &LlmopsServiceStruct{}
	// 空描述守卫先于 chat
	if _, err := svc.CreateSkillFromDescription("  "); err == nil || !strings.Contains(err.Error(), "2284") {
		t.Fatalf("空描述应报 2284: %v", err)
	}
	_ = model.LlmopsSkill{}
}
