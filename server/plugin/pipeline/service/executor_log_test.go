package service

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateLogRunes(t *testing.T) {
	short := "hello 构建日志"
	if got := truncateLogRunes(short, 16*1024); got != short {
		t.Fatalf("未超限不应截断: %q", got)
	}
	long := strings.Repeat("构建", 10000) // 40KB 中文
	got := truncateLogRunes(long, 16*1024)
	if !strings.HasSuffix(got, "\n…(截断)") {
		t.Fatalf("截断应带标记: %q", got[len(got)-20:])
	}
	body := strings.TrimSuffix(got, "\n…(截断)")
	if len(body) > 16*1024 {
		t.Fatalf("截断体超限: %d", len(body))
	}
	if !utf8.ValidString(got) {
		t.Fatal("截断应保持 UTF-8 合法（原实现字节硬切会产生乱码）")
	}
	pure := strings.Repeat("a", 20000)
	if got := truncateLogRunes(pure, 100); len(got) != 100+len("\n…(截断)") {
		t.Fatalf("ASCII 截断长度不符: %d", len(got))
	}
}

func TestReadBodySnippet(t *testing.T) {
	big := strings.Repeat("x", 5000)
	if got := readBodySnippet(strings.NewReader(big), 1024); len(got) != 1024 {
		t.Fatalf("应完整读取上限字节: %d", len(got))
	}
	small := "short body"
	if got := readBodySnippet(strings.NewReader(small), 1024); got != small {
		t.Fatalf("短内容应全读: %q", got)
	}
	if got := readBodySnippet(nil, 100); got != "" {
		t.Fatalf("nil reader 应返回空: %q", got)
	}
}
