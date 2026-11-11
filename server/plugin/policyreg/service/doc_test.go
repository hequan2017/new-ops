// Package service policyreg：制度导入单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestMaskSecretClauses 涉密条款脱敏
func TestMaskSecretClauses(t *testing.T) {
	text := "第一章 总则\n为规范科研管理制定本办法。\n机密：涉密项目清单见附件\n第二条 本办法自发布之日起施行。\n绝密：核心参数配置表\n"
	masked, count := MaskSecretClauses(text)
	if count != 2 {
		t.Fatalf("应脱敏 2 行: %d", count)
	}
	if strings.Contains(masked, "涉密项目清单") || strings.Contains(masked, "核心参数配置表") {
		t.Fatalf("涉密内容不应残留: %s", masked)
	}
	if !strings.Contains(masked, "机密★（涉密条款已自动脱敏 1）") {
		t.Fatalf("脱敏标记不符: %s", masked)
	}
	if !strings.Contains(masked, "为规范科研管理制定本办法。") {
		t.Fatal("非涉密内容应保留")
	}
	// 无涉密行
	if _, n := MaskSecretClauses("普通制度文本"); n != 0 {
		t.Fatalf("无涉密应 0: %d", n)
	}
}

// TestDocWhitelists 白名单
func TestDocWhitelists(t *testing.T) {
	for _, f := range PolicyFormats {
		if !ValidPolicyFormat(f) {
			t.Fatalf("%s 应在格式白名单", f)
		}
	}
	for _, l := range PolicySecurityLevels {
		if !ValidPolicyLevel(l) {
			t.Fatalf("%s 应在密级白名单", l)
		}
	}
	if ValidPolicyFormat("EXE") || ValidPolicyLevel("绝密") {
		t.Fatal("白名单外应拒绝")
	}
}
