// Package service 白泽容器管理：模板渲染单测（M8 C3）
package service

import "testing"

func TestRenderComposeTplBasic(t *testing.T) {
	schema := []ParamDef{
		{Key: "image", Label: "镜像", Default: "nginx:latest"},
		{Key: "port", Label: "端口", Required: true},
	}
	tpl := "services:\n  web:\n    image: {{image}}\n    ports:\n      - {{port}}:80\n"
	got, err := renderComposeTpl(tpl, schema, map[string]string{"port": "8080"})
	if err != nil {
		t.Fatalf("渲染失败: %v", err)
	}
	want := "services:\n  web:\n    image: nginx:latest\n    ports:\n      - 8080:80\n"
	if got != want {
		t.Fatalf("渲染结果不符:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderComposeTplMissingRequired(t *testing.T) {
	schema := []ParamDef{{Key: "port", Required: true}}
	if _, err := renderComposeTpl("p: {{port}}", schema, map[string]string{}); err == nil {
		t.Fatalf("必填参数缺失应报错")
	}
}

func TestRenderComposeTplUnknownPlaceholder(t *testing.T) {
	// 模板占位符未在 schema 定义 → 报错（防漏配）
	if _, err := renderComposeTpl("p: {{nope}}", nil, map[string]string{}); err == nil {
		t.Fatalf("未定义占位符应报错")
	}
}

func TestRenderComposeTplExtraParams(t *testing.T) {
	// 多余参数：schema 外传值不影响渲染（宽容，便于统一表单提交）
	got, err := renderComposeTpl("p: {{a}}", []ParamDef{{Key: "a", Default: "x"}}, map[string]string{"b": "y"})
	if err != nil {
		t.Fatalf("多余参数不应报错: %v", err)
	}
	if got != "p: x" {
		t.Fatalf("渲染结果不符: %s", got)
	}
}

func TestParseParamsSchema(t *testing.T) {
	if defs, err := ParseParamsSchema(""); err != nil || len(defs) != 0 {
		t.Fatalf("空 schema 应合法: %v %v", defs, err)
	}
	defs, err := ParseParamsSchema(`[{"key":"image","label":"镜像","default":"nginx","required":true}]`)
	if err != nil || len(defs) != 1 || defs[0].Key != "image" || !defs[0].Required {
		t.Fatalf("schema 解析不符: %v %v", defs, err)
	}
	if _, err := ParseParamsSchema(`[{"key":"a"},{"key":"a"}]`); err == nil {
		t.Fatalf("重复 key 应报错")
	}
	if _, err := ParseParamsSchema("not-json"); err == nil {
		t.Fatalf("非 JSON 应报错")
	}
}
