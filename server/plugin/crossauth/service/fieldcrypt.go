// Package service crossauth：敏感数据加密应用（需求 3.7-10）
// 口径：字段加密适配器——SM4（密钥轮换表数据密钥）/AES-GCM（asset/crypto 主密钥）
// 双算法可选，密文落库不回显；访问水印=用户+CST 时间戳嵌入返回文本（尾部标记，
// 纯函数可提取）。
package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
)

// 错误码（crossauth 段 2900-2999）
const (
	ErrCodeCaFieldParam   = 2917
	ErrCodeCaFieldNotFound = 2918
)

// CaFieldAlgorithms 算法白名单
var CaFieldAlgorithms = []string{"SM4", "AES-GCM"}

// EncryptCaField 字段加密适配器
func (s *CaService) EncryptCaField(algorithm string, plaintext []byte) (string, error) {
	valid := false
	for _, a := range CaFieldAlgorithms {
		if a == algorithm {
			valid = true
		}
	}
	if !valid {
		return "", fmt.Errorf("[%d]算法须为 SM4/AES-GCM", ErrCodeCaFieldParam)
	}
	switch algorithm {
	case "SM4":
		return s.EncryptCaData("default", plaintext)
	default: // AES-GCM 走主密钥
		enc, err := crypto.Encrypt(string(plaintext))
		if err != nil {
			return "", fmt.Errorf("[%d]AES-GCM 加密失败: %v", ErrCodeCaFieldParam, err)
		}
		return enc, nil
	}
}

// DecryptCaField 字段解密适配器
func (s *CaService) DecryptCaField(algorithm string, cipherText string) ([]byte, error) {
	switch algorithm {
	case "SM4":
		return s.DecryptCaData("default", cipherText)
	case "AES-GCM":
		plain, err := crypto.Decrypt(cipherText)
		if err != nil {
			return nil, fmt.Errorf("[%d]AES-GCM 解密失败: %v", ErrCodeCaFieldParam, err)
		}
		return []byte(plain), nil
	default:
		return nil, fmt.Errorf("[%d]算法须为 SM4/AES-GCM", ErrCodeCaFieldParam)
	}
}

// watermarkTag 水印标记
const watermarkTag = "〔CA-WM "

// WatermarkText 访问水印（纯函数，测试锚点）：用户+CST 时间戳嵌入返回文本尾部
func WatermarkText(text, userName string, now time.Time) string {
	tag := fmt.Sprintf("%s%s|%s〕", watermarkTag, userName, now.In(caCst).Format("2006-01-02 15:04:05"))
	if strings.TrimSpace(text) == "" {
		return tag
	}
	return text + "\n" + tag
}

// ParseWatermark 提取水印（纯函数，测试锚点）
func ParseWatermark(text string) (userName, ts string, ok bool) {
	idx := strings.LastIndex(text, watermarkTag)
	if idx < 0 {
		return "", "", false
	}
	rest := text[idx+len(watermarkTag):]
	end := strings.Index(rest, "〕")
	if end < 0 {
		return "", "", false
	}
	parts := strings.SplitN(rest[:end], "|", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}
