package service

import (
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/job/model"
)

func TestValidateScript(t *testing.T) {
	// 合法
	if err := validateScript(&model.JobScript{Name: "df-check", Language: "shell"}); err != nil {
		t.Fatalf("合法脚本不应报错: %v", err)
	}
	// 缺名称
	err := validateScript(&model.JobScript{Language: "shell"})
	if se, ok := err.(*scriptError); !ok || se.Code != ErrCodeScriptNameRequired {
		t.Fatalf("缺名称应报 1201: %v", err)
	}
	// 非法语言
	err = validateScript(&model.JobScript{Name: "x", Language: "batch"})
	if se, ok := err.(*scriptError); !ok || se.Code != ErrCodeScriptLangInvalid {
		t.Fatalf("非法语言应报 1203: %v", err)
	}
}

func TestValidateVariableGroup(t *testing.T) {
	// 合法变量 JSON
	if err := validateVariableGroup(&model.JobVariableGroup{
		Name:      "web-params",
		Variables: `[{"key":"port","value":"8080"}]`,
	}); err != nil {
		t.Fatalf("合法变量组不应报错: %v", err)
	}
	// 空变量允许（占位组）
	if err := validateVariableGroup(&model.JobVariableGroup{Name: "empty"}); err != nil {
		t.Fatalf("空变量组不应报错: %v", err)
	}
	// 非法 JSON
	err := validateVariableGroup(&model.JobVariableGroup{Name: "bad", Variables: `{not-json`})
	if se, ok := err.(*scriptError); !ok || se.Code != ErrCodeVariablesInvalid {
		t.Fatalf("非法 JSON 应报 1207: %v", err)
	}
}

func TestRenderTemplate(t *testing.T) {
	vars := `[{"key":"port","value":"8080"},{"key":"dir","value":"/data"}]`
	out, err := RenderTemplate("echo {{port}} > {{dir}}/app.log", vars)
	if err != nil {
		t.Fatalf("渲染不应报错: %v", err)
	}
	if out != "echo 8080 > /data/app.log" {
		t.Fatalf("渲染结果错误: %s", out)
	}
	// 未定义占位保持原样
	out, _ = RenderTemplate("keep {{unknown}}", vars)
	if out != "keep {{unknown}}" {
		t.Fatalf("未定义占位应保持原样: %s", out)
	}
	// 空变量组原样返回
	out, _ = RenderTemplate("no {{var}}", "")
	if out != "no {{var}}" {
		t.Fatalf("空变量应原样返回: %s", out)
	}
	// 非法 JSON 报错
	if _, err := RenderTemplate("x", "{bad"); err == nil {
		t.Fatal("非法变量 JSON 应报错")
	}
}
