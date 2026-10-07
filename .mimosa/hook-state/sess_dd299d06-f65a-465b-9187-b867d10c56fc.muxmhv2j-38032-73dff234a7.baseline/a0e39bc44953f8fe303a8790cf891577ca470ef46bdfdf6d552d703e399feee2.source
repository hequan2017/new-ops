package crypto

import (
	"os"
	"strings"
	"sync"
	"testing"
)

// 加解密往返测试（临时主密钥，测试后清理）
func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Setenv(MasterKeyEnv, "unit-test-master-key")
	// 重置惰性缓存
	syncReset()

	plain := "s3cret-Passw0rd-9f8e"
	enc, err := Encrypt(plain)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if strings.Contains(enc, plain) {
		t.Fatalf("密文包含明文！")
	}
	dec, err := Decrypt(enc)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if dec != plain {
		t.Fatalf("往返不一致: got %s", dec)
	}

	// 同明文两次加密 nonce 不同 → 密文不同
	enc2, _ := Encrypt(plain)
	if enc == enc2 {
		t.Fatalf("GCM nonce 应随机，两次密文不应相同")
	}
}

// 错误密钥解密必须失败
func TestDecryptWrongKey(t *testing.T) {
	t.Setenv(MasterKeyEnv, "key-a")
	syncReset()
	enc, err := Encrypt("data")
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	t.Setenv(MasterKeyEnv, "key-b")
	syncReset()
	if _, err := Decrypt(enc); err == nil {
		t.Fatalf("错误主密钥解密应失败")
	}
	os.Unsetenv(MasterKeyEnv)
	syncReset()
	if _, err := Encrypt("x"); err == nil || !strings.Contains(err.Error(), MasterKeyEnv) {
		t.Fatalf("未配置主密钥应报环境变量提示, got %v", err)
	}
}

func TestLast4(t *testing.T) {
	if got := Last4("abcdefg"); got != "defg" {
		t.Fatalf("Last4 = %s, want defg", got)
	}
	if got := Last4("ab"); got != "ab" {
		t.Fatalf("短串应原样返回, got %s", got)
	}
}

// syncReset 重置包内惰性加载缓存（测试用）
func syncReset() {
	masterKeyOnce = sync.Once{}
}
