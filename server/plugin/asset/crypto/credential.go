// Package crypto 凭据保险库加解密（AES-256-GCM）
// 主密钥只从环境变量 NEW_OPS_CREDENTIAL_MASTER_KEY 注入，源码/配置不落明文。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

// MasterKeyEnv 主密钥环境变量名
const MasterKeyEnv = "NEW_OPS_CREDENTIAL_MASTER_KEY"

var (
	masterKeyOnce sync.Once
	masterKey     []byte
	masterKeyErr  error
)

// loadMasterKey 惰性加载并派生 32 字节主密钥（SHA-256 派生，任意长度环境变量均可）
func loadMasterKey() ([]byte, error) {
	masterKeyOnce.Do(func() {
		raw := os.Getenv(MasterKeyEnv)
		if raw == "" {
			masterKeyErr = errors.New("凭据主密钥未配置：请设置环境变量 " + MasterKeyEnv)
			return
		}
		sum := sha256.Sum256([]byte(raw))
		masterKey = sum[:]
	})
	return masterKey, masterKeyErr
}

// Encrypt 用主密钥 AES-256-GCM 加密（输出 base64(nonce|cipher)）
func Encrypt(plaintext string) (string, error) {
	key, err := loadMasterKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("初始化 AES 失败: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("初始化 GCM 失败: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成 nonce 失败: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt 解密 Encrypt 的输出
func Decrypt(encoded string) (string, error) {
	key, err := loadMasterKey()
	if err != nil {
		return "", err
	}
	sealed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("密文 base64 解码失败: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("初始化 AES 失败: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("初始化 GCM 失败: %w", err)
	}
	if len(sealed) < gcm.NonceSize() {
		return "", errors.New("密文长度不合法")
	}
	nonce, ct := sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败（主密钥不匹配或密文损坏）: %w", err)
	}
	return string(plain), nil
}

// Last4 返回明文末4位（不足4位返回全量占位）
func Last4(plaintext string) string {
	r := []rune(plaintext)
	if len(r) <= 4 {
		return string(r)
	}
	return string(r[len(r)-4:])
}
