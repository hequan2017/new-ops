// Package service llmops 数据资产域纯函数单测（类型/状态白名单、LIKE 转义）
package service

import (
	"testing"
)

func TestValidDataType(t *testing.T) {
	for _, ok := range []string{"doc", "image", "audio", "video", "webpage", "qa", "structured", "semistructured"} {
		if !ValidDataType(ok) {
			t.Fatalf("合法类型 %q 被拒绝", ok)
		}
	}
	for _, bad := range []string{"", "DOC", "video;drop", "file", "docs "} {
		if ValidDataType(bad) {
			t.Fatalf("非法类型 %q 被放行", bad)
		}
	}
}

func TestValidAssetStatus(t *testing.T) {
	for _, ok := range []string{"draft", "parsed", "cataloged", "offline"} {
		if !ValidAssetStatus(ok) {
			t.Fatalf("合法状态 %q 被拒绝", ok)
		}
	}
	for _, bad := range []string{"", "DRAFT", "parsed,x", "deleted"} {
		if ValidAssetStatus(bad) {
			t.Fatalf("非法状态 %q 被放行", bad)
		}
	}
}

// TestEscapeLike LIKE 通配符按字面匹配（防 %/_ 注入扩展过滤范围）
func TestEscapeLike(t *testing.T) {
	cases := map[string]string{
		`100%`:  `100\%`,
		`a_b`:   `a\_b`,
		`x\y`:   `x\\y`,
		`%_\`:   `\%\_\\`,
		`plain`: `plain`,
	}
	for in, want := range cases {
		if got := EscapeLike(in); got != want {
			t.Fatalf("EscapeLike(%q) = %q, want %q", in, got, want)
		}
	}
}
