// Package service oplib 执行引擎单测（简化 schema 三维校验/内置执行链）
package service

import "testing"

func TestValidateArgsAgainstSchema(t *testing.T) {
	schema := `{
		"text": {"type":"string","required":true},
		"topN": {"type":"number"},
		"mode": {"type":"string","enum":["lower","upper","title"]}
	}`
	// 合法
	if err := ValidateArgsAgainstSchema(schema, map[string]any{"text": "x", "topN": 5.0, "mode": "lower"}); err != nil {
		t.Fatalf("合法参数应放行: %v", err)
	}
	// 缺必填
	if err := ValidateArgsAgainstSchema(schema, map[string]any{}); err == nil || !contains(err.Error(), "缺少必填") {
		t.Fatalf("缺必填应报: %v", err)
	}
	// 类型错
	if err := ValidateArgsAgainstSchema(schema, map[string]any{"text": 123.0}); err == nil || !contains(err.Error(), "字符串") {
		t.Fatalf("类型错应报: %v", err)
	}
	if err := ValidateArgsAgainstSchema(schema, map[string]any{"text": "x", "topN": "5"}); err == nil {
		t.Fatal("number 传字符串应拒")
	}
	// 枚举外
	if err := ValidateArgsAgainstSchema(schema, map[string]any{"text": "x", "mode": "weird"}); err == nil || !contains(err.Error(), "枚举") {
		t.Fatalf("枚举外应报: %v", err)
	}
	// 可选缺省放行
	if err := ValidateArgsAgainstSchema(schema, map[string]any{"text": "x"}); err != nil {
		t.Fatalf("可选缺失应放行: %v", err)
	}
	// 空 schema 放行
	if err := ValidateArgsAgainstSchema("", map[string]any{}); err != nil {
		t.Fatalf("空 schema 应放行: %v", err)
	}
	// 坏 schema
	if err := ValidateArgsAgainstSchema(`[1]`, map[string]any{}); err == nil {
		t.Fatal("坏 schema 应拒绝")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// TestRunBuiltin 内置执行链（参数解析失败/未注册/成功路径）
func TestRunBuiltin(t *testing.T) {
	out, err := runBuiltin("text_clean", `{"text":"  a  b  "}`)
	if err != nil || out.(string) != "a b" {
		t.Fatalf("执行异常: %v %v", out, err)
	}
	if _, err := runBuiltin("no_such_op", `{}`); err == nil || !contains(err.Error(), "未注册") {
		t.Fatalf("未注册应报: %v", err)
	}
	if _, err := runBuiltin("text_clean", `bad-json`); err == nil {
		t.Fatal("坏参数 JSON 应拒")
	}
}
