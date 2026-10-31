package service

import (
	"strings"
	"testing"

	"github.com/hequan2017/new-ops/server/plugin/dbops/model"
)

// TestWrapMagicBlock goInception magic 块组装（use 前置/选项注入/非法凭据与库名拦截）
func TestWrapMagicBlock(t *testing.T) {
	target := inceptionTarget{
		InstanceHost: "10.0.0.8", InstancePort: 3306,
		Username: "rw", Password: "p@ss", Database: "biz",
	}
	// 审核态：check=1 + 工单无 use 时前置目标库
	check, err := wrapMagicBlock(target, "update t set a=1;", false)
	if err != nil {
		t.Fatalf("wrapMagicBlock 审核态报错: %v", err)
	}
	if !strings.Contains(check, ";--check=1;*/") {
		t.Fatalf("审核态应含 check=1 选项: %q", check)
	}
	if !strings.Contains(check, "use biz;\nupdate t set a=1;") {
		t.Fatalf("无 use 工单应前置目标库: %q", check)
	}
	if !strings.Contains(check, magicStmtStart) || !strings.Contains(check, magicStmtEnd) {
		t.Fatalf("缺 magic 定界语句: %q", check)
	}
	// 执行态：execute=1 + backup + ignore-warnings；工单自带 use 不再前置
	exec, err := wrapMagicBlock(target, "use report;\ntruncate t;", true)
	if err != nil {
		t.Fatalf("wrapMagicBlock 执行态报错: %v", err)
	}
	if !strings.Contains(exec, ";--execute=1;--backup=true;--ignore-warnings=1") {
		t.Fatalf("执行态应含 execute/backup/ignore-warnings 选项: %q", exec)
	}
	if strings.Count(exec, "use ") != 1 {
		t.Fatalf("自带 use 的工单不应再前置 use: %q", exec)
	}
	// 凭据含注释选项非法字符（; */ 换行）必须拦截，防选项注释被截断逃逸
	for name, bad := range map[string]inceptionTarget{
		"密码含分号": {InstanceHost: "h", InstancePort: 3306, Username: "u", Password: "a;b", Database: "d"},
		"密码含*/": {InstanceHost: "h", InstancePort: 3306, Username: "u", Password: "a*/b", Database: "d"},
		"账号含换行": {InstanceHost: "h", InstancePort: 3306, Username: "u\nx", Password: "p", Database: "d"},
	} {
		if _, err := wrapMagicBlock(bad, "select 1;", false); err == nil {
			t.Fatalf("%s 应被拦截", name)
		}
	}
	// 库名白名单：拼接 use 前强校验（防库名夹带语句）
	if _, err := wrapMagicBlock(inceptionTarget{InstanceHost: "h", InstancePort: 3306, Username: "u", Password: "p", Database: "a;drop"}, "select 1;", false); err == nil {
		t.Fatalf("非法库名应被拦截")
	}
}

// Test_engineDSN goInception 连接 DSN 组装：超时/多语句/字符集口径固化（inception 需 multiStatements 支持魔法块定界）
func Test_engineDSN(t *testing.T) {
	got := engineDSN(&model.DbopsInceptionConfig{Host: "127.0.0.1", Port: 4000})
	want := "tcp(127.0.0.1:4000)/?timeout=10s&readTimeout=120s&writeTimeout=10s&multiStatements=true&charset=utf8mb4"
	if got != want {
		t.Fatalf("want %q got %q", want, got)
	}
}

// Test_magicOptionSafe 凭据嵌入注释选项的安全字符校验（; */ 换行会截断或逃逸 magic 块）
func Test_magicOptionSafe(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"普通密码", "p@ssw0rd-123", true},
		{"含分号拒绝", "pass;drop", false},
		{"含注释符拒绝", "pass*/x", false},
		{"含换行拒绝", "pass\nline", false},
		{"含回车拒绝", "pass\rline", false},
	}
	for _, c := range cases {
		if got := magicOptionSafe(c.in); got != c.want {
			t.Fatalf("%s: want %v got %v", c.name, c.want, got)
		}
	}
}
