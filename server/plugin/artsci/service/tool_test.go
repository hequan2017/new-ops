// Package service artsci：工具注册表纯函数单测（schema 校验/学科白名单/分发；DB 依赖路径不覆盖）
package service

import (
	"strings"
	"testing"
)

func TestValidateArgsAgainstSchema(t *testing.T) {
	schema := `{"epsilon":{"type":"number","required":true},"name":{"type":"string","enum":["a","b"]},"flag":{"type":"bool"}}`
	cases := []struct {
		name    string
		schema  string
		args    map[string]any
		wantErr string
	}{
		{"空schema放行", "", map[string]any{}, ""},
		{"全部合法", schema, map[string]any{"epsilon": 0.5, "name": "a", "flag": true}, ""},
		{"缺必填", schema, map[string]any{"name": "a"}, "缺少必填参数: epsilon"},
		{"数字类型错", schema, map[string]any{"epsilon": "x"}, "epsilon 须为数字"},
		{"字符串类型错", schema, map[string]any{"epsilon": 1.0, "name": 3}, "name 须为字符串"},
		{"枚举外", schema, map[string]any{"epsilon": 1.0, "name": "c"}, "不在枚举内"},
		{"schema坏JSON", "{bad", map[string]any{}, "schema 解析失败"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateArgsAgainstSchema(tc.schema, tc.args)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("期望通过，实际: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("期望含 %q，实际: %v", tc.wantErr, err)
			}
		})
	}
}

func TestDispatch(t *testing.T) {
	RegisterModule("test-mod", func(args map[string]any) (any, error) {
		return map[string]any{"echo": args["text"]}, nil
	})
	t.Cleanup(func() { delete(builtinModules, "test-mod") })

	s := &ArtsciService{}
	// 命中注册表：校验→执行
	got, err := s.dispatch("test-mod", `{"text":{"type":"string","required":true}}`, `{"text":"hi"}`)
	if err != nil {
		t.Fatalf("期望成功，实际: %v", err)
	}
	if m, _ := got.(map[string]any); m["echo"] != "hi" {
		t.Fatalf("回显不符: %v", got)
	}
	// 未注册 module → 2505
	_, err = s.dispatch("no-such-module", "", `{}`)
	if err == nil || !strings.Contains(err.Error(), "[2505]") {
		t.Fatalf("期望 2505 未注册，实际: %v", err)
	}
	// 参数坏 JSON → 2507
	_, err = s.dispatch("test-mod", "", "{bad")
	if err == nil || !strings.Contains(err.Error(), "[2507]") {
		t.Fatalf("期望 2507 参数解析失败，实际: %v", err)
	}
	// schema 校验不过不触达实现
	_, err = s.dispatch("test-mod", `{"text":{"type":"string","required":true}}`, `{}`)
	if err == nil || !strings.Contains(err.Error(), "缺少必填参数") {
		t.Fatalf("期望必填校验失败，实际: %v", err)
	}
}

func TestSubjectWhitelist(t *testing.T) {
	for _, s := range []string{"数学", "文科", "图像"} {
		if !ValidSubject(s) {
			t.Fatalf("%s 应在白名单", s)
		}
	}
	if ValidSubject("英语") || ValidSubject("") {
		t.Fatal("白名单外应拒绝")
	}
}

func TestTruncateRunText(t *testing.T) {
	long := strings.Repeat("x", 600)
	if got := truncateRunText(long); len(got) != 500 {
		t.Fatalf("期望截断到 500，实际 %d", len(got))
	}
	if got := truncateRunText("ok"); got != "ok" {
		t.Fatalf("短文本应原样: %v", got)
	}
}
