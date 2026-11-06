// Package service llmops 数据源接入纯函数单测（类型/表名白名单、DSN 拼接）
package service

import (
	"strings"
	"testing"
)

func TestValidSourceType(t *testing.T) {
	for _, ok := range []string{"mysql", "postgres"} {
		if !ValidSourceType(ok) {
			t.Fatalf("合法类型 %q 被拒绝", ok)
		}
	}
	for _, bad := range []string{"", "MySQL", "sqlite", "mssql", "mysql;"} {
		if ValidSourceType(bad) {
			t.Fatalf("非法类型 %q 被放行", bad)
		}
	}
}

// TestValidTableName 表名白名单（预览接口防 SQL 注入）
func TestValidTableName(t *testing.T) {
	for _, ok := range []string{"users", "t_001", "A", "aB_9"} {
		if !ValidTableName(ok) {
			t.Fatalf("合法表名 %q 被拒绝", ok)
		}
	}
	for _, bad := range []string{"", "user;drop", "t-1", "`t`", "a b", strings.Repeat("x", 65), "1'1"} {
		if ValidTableName(bad) {
			t.Fatalf("非法表名 %q 被放行", bad)
		}
	}
}

func TestBuildDSN(t *testing.T) {
	m := &DataSourceIn{Type: "mysql", Host: "10.0.0.1", Port: 3306, Database: "shop", User: "u1", Password: "p@ss"}
	dsn := BuildDSN(m)
	for _, want := range []string{"u1:p@ss@tcp(10.0.0.1:3306)/shop", "charset=utf8mb4", "timeout=5s"} {
		if !strings.Contains(dsn, want) {
			t.Fatalf("mysql DSN 缺 %q: %s", want, dsn)
		}
	}
	p := &DataSourceIn{Type: "postgres", Host: "pg.local", Port: 5432, Database: "db", User: "u2", Password: "pp"}
	pdsn := BuildDSN(p)
	for _, want := range []string{"host=pg.local", "port=5432", "user=u2", "password=pp", "dbname=db", "sslmode=disable"} {
		if !strings.Contains(pdsn, want) {
			t.Fatalf("postgres DSN 缺 %q: %s", want, pdsn)
		}
	}
}

// TestOpenDSRejectsType 非白名单类型在 openDS 入口即拒（不触网络）
func TestOpenDSRejectsType(t *testing.T) {
	if _, err := openDS(&DataSourceIn{Type: "oracle"}); err == nil || !strings.Contains(err.Error(), "2212") {
		t.Fatalf("非法类型应报 2212: %v", err)
	}
}
