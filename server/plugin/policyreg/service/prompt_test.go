// Package service policyreg：提示词与问答单测（纯函数+假模型，无 DB）
package service

import (
	"context"
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/policyreg/model"
)

// TestRenderPrTemplate 渲染成功/缺失变量
func TestRenderPrTemplate(t *testing.T) {
	out, err := renderPrTemplate("请回答 {{question}}（角色 {{role}}）", map[string]string{"question": "报销流程", "role": "教师"})
	if err != nil || out != "请回答 报销流程（角色 教师）" {
		t.Fatalf("渲染不符: %s %v", out, err)
	}
	if _, err := renderPrTemplate("引用 {{refs}}", map[string]string{}); err == nil || !strings.Contains(err.Error(), "缺失占位变量") {
		t.Fatalf("缺失变量应报错: %v", err)
	}
	if vars := parsePrVars("{{a}} 与 {{b}} 再见 {{a}}"); len(vars) != 2 || vars[0] != "a" || vars[1] != "b" {
		t.Fatalf("解析应去重保序: %v", vars)
	}
}

// TestRulePolicyAnswer 规则口径回答
func TestRulePolicyAnswer(t *testing.T) {
	docs := []model.PolicyDoc{
		{Name: "科研经费管理办法", SecurityLevel: "内部", Format: "TXT", RawText: "横向课题经费实行专账核算，预算调整须审批。"},
	}
	answer := rulePolicyAnswer("横向经费怎么调整预算", docs)
	if !strings.Contains(answer, "1 项相关制度") || !strings.Contains(answer, "科研经费管理办法") || !strings.Contains(answer, "专账核算") {
		t.Fatalf("规则回答应含命中制度: %s", answer)
	}
	empty := rulePolicyAnswer("无关问题", nil)
	if !strings.Contains(empty, "未检索到") {
		t.Fatalf("空命中应兜底: %s", empty)
	}
}

// TestPolicyChatFnInjection 假模型注入锚点（httptest 免外部服务）
func TestPolicyChatFnInjection(t *testing.T) {
	orig := policyChatFn
	defer func() { policyChatFn = orig }()
	policyChatFn = func(ctx context.Context, modelId uint, prompt string) (string, error) {
		if modelId != 42 {
			t.Fatalf("modelId 应透传: %d", modelId)
		}
		if !strings.Contains(prompt, "制度问答") {
			t.Fatalf("prompt 应含模板注入: %s", prompt)
		}
		return "【假模型】依据《科研经费管理办法》第 5 条，预算调整须经审批。", nil
	}
	out, err := policyChatFn(context.Background(), 42, "你是高校制度问答助手。制度问答模板注入测试。")
	if err != nil || out == "" {
		t.Fatalf("假模型应返回: %s %v", out, err)
	}
}
