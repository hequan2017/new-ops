// Package service crossauth：MFA 单测（纯函数，RFC 6238 官方向量，无 DB）
package service

import (
	"testing"
)

// rfcSecret RFC 6238 测试密钥 "12345678901234567890"
func rfcSecret() []byte {
	return []byte("12345678901234567890")
}

// TestTotpRfc6238 RFC 6238 SHA-1 官方向量（T=59/1111111109/1234567890）
func TestTotpRfc6238(t *testing.T) {
	cases := map[int64]string{
		59:          "287082",
		1111111109:  "081804",
		1234567890:  "005924",
		2000000000:  "279037",
	}
	for ts, want := range cases {
		if got := totpCode(rfcSecret(), ts, 30); got != want {
			t.Fatalf("T=%d 应 %s 得 %s", ts, want, got)
		}
	}
}

// TestVerifyTotp 窗口容差
func TestVerifyTotp(t *testing.T) {
	now := int64(59)
	current := totpCode(rfcSecret(), now, 30)
	if !verifyTotp(rfcSecret(), current, now) {
		t.Fatal("当前窗口码应通过")
	}
	// 上/下一窗口容差
	if !verifyTotp(rfcSecret(), totpCode(rfcSecret(), now+30, 30), now) {
		t.Fatal("下一窗口码应容差通过")
	}
	if verifyTotp(rfcSecret(), totpCode(rfcSecret(), now+120, 30), now) {
		t.Fatal("两窗外应拒绝")
	}
	if verifyTotp(rfcSecret(), "000000", now) {
		t.Fatal("错码应拒绝")
	}
}
