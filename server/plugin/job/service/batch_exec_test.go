package service

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/hequan2017/new-ops/server/plugin/job/model"
)

func TestTruncateRunes(t *testing.T) {
	if got := truncateRunes("hello", 10); got != "hello" {
		t.Fatalf("未超限不应截断: %q", got)
	}
	long := strings.Repeat("好", 40000) // 120KB
	got := truncateRunes(long, 60000)
	if len(got) > 60000 {
		t.Fatalf("截断后仍超限: %d", len(got))
	}
	if !utf8.ValidString(got) {
		t.Fatal("截断应保持 UTF-8 合法（字符边界）")
	}
}

func TestBoundedWriter(t *testing.T) {
	w := &boundedWriter{max: 1024}
	for i := 0; i < 100; i++ {
		if _, err := w.Write([]byte(strings.Repeat("a", 4096))); err != nil {
			t.Fatalf("写入失败: %v", err)
		}
		if len(w.buf) > 1024 {
			t.Fatalf("缓冲超上限: %d", len(w.buf))
		}
	}
	if len(w.Bytes()) != 1024 {
		t.Fatalf("应保留头部: %d", len(w.Bytes()))
	}
}

func TestValidResultStatus(t *testing.T) {
	for _, s := range []string{model.ResultSuccess, model.ResultFailed, model.ResultTimeout, model.ResultCanceled, model.ResultSkipped} {
		if !validResultStatus(s) {
			t.Fatalf("模型常量 %q 应为合法过滤值", s)
		}
	}
	for _, s := range []string{"", "success", "成功 ", "任意"} {
		if validResultStatus(s) {
			t.Fatalf("%q 不应通过校验", s)
		}
	}
}
