// Package service fanhua：方言问答单测（纯函数+假模型，无 DB）
package service

import (
	"context"
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/fanhua/model"
)

// TestBuildQaPrompt 注入与助老口径
func TestBuildQaPrompt(t *testing.T) {
	refs := []paraHit{{FanhuaParallel: model.FanhuaParallel{Sh: "侬好", Pu: "你好"}}}
	p := buildQaPrompt("打招呼怎么说", refs, false)
	if !strings.Contains(p, "侬好") || !strings.Contains(p, "打招呼怎么说") {
		t.Fatalf("应注入对照语料与问题: %s", p)
	}
	p2 := buildQaPrompt("问", nil, true)
	if !strings.Contains(p2, "助老模式") {
		t.Fatalf("助老口径应注入: %s", p2)
	}
}

// TestRuleQaAnswer 规则口径组合
func TestRuleQaAnswer(t *testing.T) {
	refs := []paraHit{{FanhuaParallel: model.FanhuaParallel{Sh: "侬好", Pu: "你好", En: "Hello"}}}
	a := ruleQaAnswer("问", refs)
	for _, want := range []string{"侬好", "你好", "Hello"} {
		if !strings.Contains(a, want) {
			t.Fatalf("应含 %s: %s", want, a)
		}
	}
	if a2 := ruleQaAnswer("问", nil); !strings.Contains(a2, "暂无对照语料") {
		t.Fatalf("空语料口径: %s", a2)
	}
}

// TestNoopTranslate 翻译适配器契约
func TestNoopTranslate(t *testing.T) {
	if _, ok := Translate.(NoopTranslateAdapter); !ok {
		t.Fatal("默认应为 NoopTranslateAdapter")
	}
	_, err := Translate.Translate(context.Background(), "text", "en")
	if err == nil || !strings.Contains(err.Error(), "[2627]") {
		t.Fatalf("未配置应报 2627: %v", err)
	}
}

// TestQaPromptElderPrefix 助老口径与 chatFn 注入锚点
func TestQaPromptElderPrefix(t *testing.T) {
	orig := qaChatFn
	defer func() { qaChatFn = orig }()
	var gotPrompt string
	qaChatFn = func(ctx context.Context, modelId uint, prompt string) (string, error) {
		gotPrompt = prompt
		return "侬好！普通话：你好。", nil
	}
	// elder 口径在 prompt 层生效（buildQaPrompt），此处验证注入锚点可替换且默认未被误触
	p := buildQaPrompt("问", nil, true)
	if !strings.Contains(p, "助老模式") || gotPrompt != "" {
		t.Fatalf("prompt 应含助老口径且 chatFn 未被误触: %q %q", p, gotPrompt)
	}
}
