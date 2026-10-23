package service

import (
	"testing"
	"time"
)

func TestFilePushTimeout(t *testing.T) {
	if got := filePushTimeout(0); got != 60*time.Second {
		t.Fatalf("空文件时限应为 60s: %v", got)
	}
	if got := filePushTimeout(1 << 30); got != 30*time.Minute { // 1GB → 60s+4096s 超上限，钳制
		t.Fatalf("大文件时限应钳制 30min: %v", got)
	}
	mid := filePushTimeout(256 * 1024 * 10) // 2.5MB → 60s+10s
	if mid != 70*time.Second {
		t.Fatalf("小文件时限线性计算: %v", mid)
	}
}

func TestValidateFileMode(t *testing.T) {
	for _, m := range []string{"", "0644", "644", "0755", "0600"} {
		if _, err := validateFileMode(m); err != nil {
			t.Fatalf("mode %q 应合法: %v", m, err)
		}
	}
	for _, m := range []string{"0999", "abc", "0000", "17777777"} {
		if _, err := validateFileMode(m); err == nil {
			t.Fatalf("mode %q 应拒绝", m)
		}
	}
}
