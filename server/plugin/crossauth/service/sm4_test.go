// Package service crossauth：SM4 单测（GB/T 32907 标准向量，无 DB）
package service

import (
	"bytes"
	"encoding/hex"
	"testing"
	"time"
)

// TestSm4Gb32907Vector GB/T 32907 标准测试向量
func TestSm4Gb32907Vector(t *testing.T) {
	key, _ := hex.DecodeString("0123456789abcdeffedcba9876543210")
	plain, _ := hex.DecodeString("0123456789abcdeffedcba9876543210")
	rk := sm4ExpandKeys(key)
	enc := sm4EncryptBlock(rk, plain)
	want, _ := hex.DecodeString("681edf34d206965e86b3e94f536e4246")
	if !bytes.Equal(enc, want) {
		t.Fatalf("标准向量不符: got %x want %x", enc, want)
	}
}

// TestSm4Roundtrip ECB 加解密回环（多块+PKCS#7）
func TestSm4Roundtrip(t *testing.T) {
	key, _ := hex.DecodeString("fedcba98765432100123456789abcdef")
	msgs := []string{"", "a", "跨校安全平台 SM4 加密", "0123456789abcdef0123456789abcdef0123456789abcdef"}
	for _, msg := range msgs {
		enc, err := SM4EncryptECB(key, []byte(msg))
		if err != nil {
			t.Fatalf("加密失败: %v", err)
		}
		dec, err := SM4DecryptECB(key, enc)
		if err != nil {
			t.Fatalf("解密失败: %v", err)
		}
		if string(dec) != msg {
			t.Fatalf("回环不符: %q != %q", dec, msg)
		}
	}
	// 密钥长度白名单
	if _, err := SM4EncryptECB([]byte("short"), []byte("x")); err == nil {
		t.Fatal("非 16 字节密钥应报错")
	}
}

// TestNextRotation 90 天周期
func TestNextRotation(t *testing.T) {
	rotated := time.Date(2026, 1, 1, 10, 0, 0, 0, time.Local)
	want := rotated.AddDate(0, 0, 90)
	if got := nextRotation(rotated, 90); !got.Equal(want) {
		t.Fatalf("90 天应 %v 得 %v", want, got)
	}
	// 缺省 90
	if got := nextRotation(rotated, 0); !got.Equal(want) {
		t.Fatalf("缺省应 90 天: %v", got)
	}
	if got := nextRotation(rotated, 30); got.Sub(rotated) != 30*24*time.Hour {
		t.Fatalf("30 天周期应可配: %v", got)
	}
}
