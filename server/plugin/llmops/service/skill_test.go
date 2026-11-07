// Package service llmops 技能域单测（定义解析/模板渲染/适配器执行）
package service

import (
	"context"
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/llmops/model"
)

func TestParseSkillDef(t *testing.T) {
	def, err := ParseSkillDef(`{"kind":"prompt","system":"sys","template":"翻译 {input} 成 {lang}","params":{"lang":{"default":"英文"}}}`)
	if err != nil || def.Kind != "prompt" || len(def.Params) != 1 {
		t.Fatalf("合法定义异常: %+v %v", def, err)
	}
	if _, err := ParseSkillDef(`{"kind":"code","template":"x"}`); err == nil {
		t.Fatal("非 prompt 型应拒绝")
	}
	if _, err := ParseSkillDef(`{"kind":"prompt"}`); err == nil {
		t.Fatal("缺模板应拒绝")
	}
	if _, err := ParseSkillDef(`bad`); err == nil {
		t.Fatal("坏 JSON 应拒绝")
	}
}

func TestRenderSkillTemplate(t *testing.T) {
	got := renderSkillTemplate("把 {input} 翻译成 {lang}", "你好", map[string]string{"lang": "英文"})
	if got != "把 你好 翻译成 英文" {
		t.Fatalf("渲染异常: %q", got)
	}
	// 未提供参数保留占位
	got = renderSkillTemplate("表：{schema} 需求：{input}", "查订单", nil)
	if got != "表：{schema} 需求：查订单" {
		t.Fatalf("缺参应保留占位: %q", got)
	}
}

// TestSkillAdapter_Execute 适配器执行链（fake chat 注入）
func TestSkillAdapter_Execute(t *testing.T) {
	var sysMsg, userMsg string
	fn := func(_ string, msgs []ChatMessage) (string, error) {
		for _, m := range msgs {
			if m.Role == "system" {
				sysMsg = m.Content
			}
			if m.Role == "user" {
				userMsg = m.Content
			}
		}
		return "译文", nil
	}
	sk := model.LlmopsSkill{
		Name: "中英翻译", Desc: "翻译",
		DefJSON: `{"kind":"prompt","system":"你是翻译","template":"翻译成{lang}：{input}","params":{"lang":{"default":"英文"}}}`,
	}
	ad := &skillToolAdapter{skill: sk, fn: fn}
	if ad.Name() != "skill:中英翻译" {
		t.Fatalf("工具名异常: %q", ad.Name())
	}
	// 纯文本输入
	out, err := ad.Execute(context.Background(), "你好世界")
	if err != nil || out != "译文" || userMsg != "翻译成英文：你好世界" || sysMsg != "你是翻译" {
		t.Fatalf("纯文本执行异常: %q %q %v", userMsg, sysMsg, err)
	}
	// JSON 包装（带参数覆盖）
	out, err = ad.Execute(context.Background(), `{"input":"晚上好","params":{"lang":"日语"}}`)
	if err != nil || userMsg != "翻译成日语：晚上好" {
		t.Fatalf("JSON 包装执行异常: %q %v", userMsg, err)
	}
	_ = strings.Contains
}
