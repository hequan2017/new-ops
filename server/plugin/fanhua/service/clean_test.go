// Package service fanhua：清洗流水线单测（纯函数，无 DB）
package service

import (
	"strings"
	"testing"
)

// TestCleanTextMask 脱敏：手机号/身份证中间打码
func TestCleanTextMask(t *testing.T) {
	in := "联系 13812345678，身份证 310101199001011234，拒收电话 19999999999。"
	out, stats := CleanText(in, false)
	if stats.MaskedPhone != 2 {
		t.Fatalf("应命中 2 个手机号: %d", stats.MaskedPhone)
	}
	if stats.MaskedID != 1 {
		t.Fatalf("应命中 1 个身份证: %d", stats.MaskedID)
	}
	if !strings.Contains(out, "138****5678") {
		t.Fatalf("手机号脱敏格式不符: %s", out)
	}
	if !strings.Contains(out, "310101********1234") {
		t.Fatalf("身份证脱敏格式不符: %s", out)
	}
	if strings.Contains(out, "13812345678") || strings.Contains(out, "310101199001011234") {
		t.Fatal("明文不应残留")
	}
}

// TestCleanTextZeroWidth 零宽字符剔除
func TestCleanTextZeroWidth(t *testing.T) {
	in := "沪语\u200B语料\uFEFF测试"
	out, stats := CleanText(in, false)
	if stats.RemovedZeroWidth != 2 {
		t.Fatalf("应剔除 2 个零宽字符: %d", stats.RemovedZeroWidth)
	}
	if strings.ContainsAny(out, "\u200B\uFEFF") {
		t.Fatalf("零宽字符应剔除干净: %q", out)
	}
}

// TestCleanTextGarbled 乱码率检测
func TestCleanTextGarbled(t *testing.T) {
	garbled := strings.Repeat("\uFFFD", 5) + "两字"
	_, stats := CleanText(garbled, false)
	if !stats.Garbled {
		t.Fatal("替换符占比 >30% 应判乱码")
	}
	_, stats2 := CleanText("正常沪语语料，只有\uFFFD一个", false)
	if stats2.Garbled {
		t.Fatal("低占比不应判乱码")
	}
}

// TestDedupeLines 重复行剔除
func TestDedupeLines(t *testing.T) {
	in := "第一行\n第二行\n 第一行 \n\n第三行\n第二行"
	out, removed := dedupeLines(in)
	if removed != 2 {
		t.Fatalf("应剔除 2 行: %d", removed)
	}
	if strings.Count(out, "第一行") != 1 || strings.Count(out, "第二行") != 1 {
		t.Fatalf("重复行应仅保留首次: %q", out)
	}
}

// TestBatchCleanTexts 批处理报告
func TestBatchCleanTexts(t *testing.T) {
	reports := BatchCleanTexts([]string{"文本A\u200B", "文本B"}, true)
	if len(reports) != 2 {
		t.Fatalf("报告数不符: %d", len(reports))
	}
	if reports[0].Index != 0 || reports[0].Stats.RemovedZeroWidth != 1 {
		t.Fatalf("逐条统计不符: %+v", reports[0])
	}
}
