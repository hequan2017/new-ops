// Package service crossauth：字段加密与水印单测（SM4 走 DB 需跳过，AES-GCM/水印纯函数）
package service

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestMain 注入凭据主密钥（asset/crypto sync.Once 惰性加载，须先于首个加密调用）
func TestMain(m *testing.M) {
	_ = os.Setenv("NEW_OPS_CREDENTIAL_MASTER_KEY", "unit-test-master-key")
	os.Exit(m.Run())
}

// TestEncryptCaFieldAesGcm AES-GCM 适配器回环（主密钥，无 DB）
func TestEncryptCaFieldAesGcm(t *testing.T) {
	s := &CaService{}
	enc, err := s.EncryptCaField("AES-GCM", []byte("跨校敏感字段 6222021234567890"))
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if strings.Contains(enc, "622202") {
		t.Fatal("密文不应包含明文片段")
	}
	dec, err := s.DecryptCaField("AES-GCM", enc)
	if err != nil || string(dec) != "跨校敏感字段 6222021234567890" {
		t.Fatalf("回环不符: %s %v", dec, err)
	}
	if _, err := s.EncryptCaField("RSA", []byte("x")); err == nil {
		t.Fatal("算法白名单外应报错")
	}
}

// TestWatermarkText 水印嵌入与提取
func TestWatermarkText(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 30, 0, 0, time.Local)
	wm := WatermarkText("论文查重报告正文", "zhangsan", now)
	if !strings.Contains(wm, "〔CA-WM zhangsan|2026-11-11 10:30:00〕") {
		t.Fatalf("水印标记不符: %s", wm)
	}
	user, ts, ok := ParseWatermark(wm)
	if !ok || user != "zhangsan" || ts != "2026-11-11 10:30:00" {
		t.Fatalf("水印提取不符: %s %s %v", user, ts, ok)
	}
	// 空文本
	wm2 := WatermarkText("", "lisi", now)
	if !strings.HasPrefix(wm2, "〔CA-WM lisi|") {
		t.Fatalf("空文本应仅水印: %s", wm2)
	}
	// 无水印
	if _, _, ok := ParseWatermark("普通文本"); ok {
		t.Fatal("无水印应 false")
	}
}
