package ast

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 原测试依赖本仓库已移除的 plugin/gva 示例插件文件且运行时会改写仓库源码（永不通过），
// 改为 t.TempDir() 封闭测试：临时 enter.go 上跑 Parse→Injection→Format 断言，Rollback 反向验证。

const (
	hermeticServiceEnterSrc = `package service

type ServiceGroup struct {
	Old Old
}
`
	hermeticApiEnterSrc = `package api

import (
	"github.com/hequan2017/new-ops/server/plugin/gva/service"
)

type ApiGroup struct {
	Old Old
}

var (
	OldApi = api.Old
)
`
)

func writeTempEnter(t *testing.T, name, src string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(src), 0644); err != nil {
		t.Fatalf("写临时 enter 失败: %v", err)
	}
	return p
}

func TestPluginEnter_Injection(t *testing.T) {
	t.Run("service enter 注入结构体字段", func(t *testing.T) {
		p := writeTempEnter(t, "service_enter.go", hermeticServiceEnterSrc)
		a := &PluginEnter{
			Type: TypePluginServiceEnter, Path: p,
			StructName: "User", StructCamelName: "user",
		}
		file, err := a.Parse(p, nil)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if err := a.Injection(file); err != nil {
			t.Fatalf("Injection: %v", err)
		}
		if err := a.Format(p, nil, file); err != nil {
			t.Fatalf("Format: %v", err)
		}
		got, _ := os.ReadFile(p)
		if !strings.Contains(string(got), "User user") {
			t.Fatalf("注入后缺字段 User user:\n%s", got)
		}
	})

	t.Run("api enter 注入字段+变量+import", func(t *testing.T) {
		p := writeTempEnter(t, "api_enter.go", hermeticApiEnterSrc)
		a := &PluginEnter{
			Type: TypePluginApiEnter, Path: p,
			ImportPath:      `"github.com/hequan2017/new-ops/server/plugin/gva/service"`,
			StructName:      "User",
			StructCamelName: "user",
			ModuleName:      "serviceUser",
			GroupName:       "Service",
			PackageName:     "service",
			ServiceName:     "User",
		}
		file, err := a.Parse(p, nil)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if err := a.Injection(file); err != nil {
			t.Fatalf("Injection: %v", err)
		}
		if err := a.Format(p, nil, file); err != nil {
			t.Fatalf("Format: %v", err)
		}
		got, _ := os.ReadFile(p)
		for _, want := range []string{"User user", "serviceUser", "service.Service.User"} {
			if !strings.Contains(string(got), want) {
				t.Fatalf("注入后缺 %q:\n%s", want, got)
			}
		}
	})

	t.Run("Parse 失败后 Injection 不 panic", func(t *testing.T) {
		a := &PluginEnter{Type: TypePluginApiEnter, Path: "/nonexistent/enter.go", StructName: "User"}
		if _, err := a.Parse(a.Path, nil); err == nil {
			t.Fatalf("缺失文件 Parse 应报错")
		}
		if err := a.Injection(nil); err == nil {
			t.Fatalf("nil file 注入应返回错误而非 panic")
		}
		if err := a.Rollback(nil); err == nil {
			t.Fatalf("nil file 回滚应返回错误而非 panic")
		}
	})
}

func TestPluginEnter_Rollback(t *testing.T) {
	t.Run("api enter 回滚字段与变量", func(t *testing.T) {
		p := writeTempEnter(t, "api_enter.go", hermeticApiEnterSrc)
		a := &PluginEnter{
			Type: TypePluginApiEnter, Path: p,
			ImportPath:      `"github.com/hequan2017/new-ops/server/plugin/gva/service"`,
			StructName:      "User",
			StructCamelName: "user",
			ModuleName:      "serviceUser",
			GroupName:       "Service",
			PackageName:     "service",
			ServiceName:     "User",
		}
		file, err := a.Parse(p, nil)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if err := a.Injection(file); err != nil {
			t.Fatalf("Injection: %v", err)
		}
		if err := a.Format(p, nil, file); err != nil {
			t.Fatalf("Format: %v", err)
		}
		// 回滚注入的内容
		file, err = a.Parse(p, nil)
		if err != nil {
			t.Fatalf("重 Parse: %v", err)
		}
		if err := a.Rollback(file); err != nil {
			t.Fatalf("Rollback: %v", err)
		}
		if err := a.Format(p, nil, file); err != nil {
			t.Fatalf("Format: %v", err)
		}
		got, _ := os.ReadFile(p)
		for _, gone := range []string{"User user", "serviceUser", "service.Service.User"} {
			if strings.Contains(string(got), gone) {
				t.Fatalf("回滚后仍残留 %q:\n%s", gone, got)
			}
		}
		if !strings.Contains(string(got), "OldApi") {
			t.Fatalf("回滚误删了原有变量:\n%s", got)
		}
	})
}
