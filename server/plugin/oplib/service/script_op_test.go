// Package service oplib 脚本算子单测（python3 实执行：正常/超时/坏脚本/守卫）
package service

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/oplib/model"
)

func TestValidScriptLang(t *testing.T) {
	if !ValidScriptLang("python3") || !ValidScriptLang("node") {
		t.Fatal("白名单应放行 python3/node")
	}
	for _, bad := range []string{"", "sh", "bash", "perl", "python"} {
		if ValidScriptLang(bad) {
			t.Fatalf("非白名单 %q 应拒绝", bad)
		}
	}
}

func mkScriptOp(lang, code string) *model.OplibOperator {
	return &model.OplibOperator{ImplType: "script", ScriptLang: lang, ScriptCode: code}
}

func TestSaveScriptFields(t *testing.T) {
	if err := saveScriptFields(&model.OplibOperator{ImplType: "builtin"}); err != nil {
		t.Fatalf("builtin 不校验: %v", err)
	}
	if err := saveScriptFields(mkScriptOp("sh", "x")); err == nil {
		t.Fatal("sh 应拒绝")
	}
	if err := saveScriptFields(mkScriptOp("python3", "  ")); err == nil {
		t.Fatal("空代码应拒绝")
	}
	if err := saveScriptFields(mkScriptOp("python3", strings.Repeat("x", 65<<10))); err == nil {
		t.Fatal("超 64KB 应拒绝")
	}
}

func TestRunScriptOperator_Python3(t *testing.T) {
	if _, err := execLookPath("python3"); err != nil {
		t.Skip("python3 不可用，跳过实执行")
	}
	// 正常：输出 JSON 契约
	op := mkScriptOp("python3", "import json,sys\nargs=json.load(open(sys.argv[1]))\nprint(json.dumps({'echo': args['text'], 'lang': 'py'}))")
	out, err := RunScriptOperator(op, `{"text":"你好"}`)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	m := out.(map[string]any)
	if m["echo"] != "你好" {
		t.Fatalf("参数传递异常: %v", m)
	}
	// 非 JSON 输出 → 原文返回
	op2 := mkScriptOp("python3", "print('plain output')")
	out2, err := RunScriptOperator(op2, `{}`)
	if err != nil || out2.(string) != "plain output" {
		t.Fatalf("非 JSON 输出应原文返回: %v %v", out2, err)
	}
	// 失败脚本：stderr 回传
	op3 := mkScriptOp("python3", "raise SystemExit('boom-reason')")
	if _, err := RunScriptOperator(op3, `{}`); err == nil || !strings.Contains(err.Error(), "boom-reason") {
		t.Fatalf("失败脚本应回传错误: %v", err)
	}
	// 超时强杀（注入收窄超时窗口）
	oldTimeout := scriptTimeout
	scriptTimeout = 300 * time.Millisecond
	defer func() { scriptTimeout = oldTimeout }()
	op4 := mkScriptOp("python3", "import time\ntime.sleep(2)")
	if _, err := RunScriptOperator(op4, `{}`); err == nil || !strings.Contains(err.Error(), "超时") {
		t.Fatalf("超时应报: %v", err)
	}
}

func execLookPath(name string) (string, error) {
	return exec.LookPath(name)
}


