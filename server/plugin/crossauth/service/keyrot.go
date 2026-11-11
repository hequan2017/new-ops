// Package service crossauth：密钥轮换（需求 3.7-10）
// 口径：90 天周期（CycleDays 可配，缺省 90）；数据密钥 16 字节随机生成、AES-GCM
// 主密钥密文落库不回显；轮换=生成新密钥+推算下次轮换时间（纯函数 nextRotation）。
package service

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// caCst 东八区
var caCst = time.FixedZone("CST", 8*3600)

// nextRotation 下次轮换时间（纯函数，测试锚点）：rotatedAt + cycleDays
func nextRotation(rotatedAt time.Time, cycleDays int) time.Time {
	if cycleDays < 1 {
		cycleDays = 90
	}
	return rotatedAt.AddDate(0, 0, cycleDays)
}

// SaveCaKeyRotation 登记/轮换密钥（按密钥名幂等；生成 16 字节随机数据密钥）
func (s *CaService) SaveCaKeyRotation(keyName string, cycleDays int) (*model.CaKeyRotation, error) {
	if strings.TrimSpace(keyName) == "" {
		return nil, fmt.Errorf("[%d]密钥名必填", ErrCodeCaSm4Param)
	}
	if cycleDays < 1 {
		cycleDays = 90
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	enc, err := crypto.Encrypt(string(raw))
	if err != nil {
		return nil, fmt.Errorf("[%d]密钥加密失败: %v", ErrCodeCaSm4Param, err)
	}
	now := time.Now().In(caCst)
	next := nextRotation(now, cycleDays)
	var exist model.CaKeyRotation
	if err := global.GVA_DB.Where("key_name = ?", keyName).First(&exist).Error; err == nil {
		if err := global.GVA_DB.Model(&exist).Updates(map[string]any{
			"secret_enc": enc, "cycle_days": cycleDays,
			"rotated_at": now.Format("2006-01-02 15:04"), "next_rotation_at": next.In(caCst).Format("2006-01-02 15:04"),
		}).Error; err != nil {
			return nil, err
		}
		exist.RotatedAt = now.Format("2006-01-02 15:04")
		return &exist, nil
	}
	rec := model.CaKeyRotation{
		KeyName: keyName, Algorithm: "SM4-ECB", SecretEnc: enc, CycleDays: cycleDays,
		RotatedAt: now.Format("2006-01-02 15:04"), NextRotationAt: next.In(caCst).Format("2006-01-02 15:04"),
		Status: "启用",
	}
	if err := global.GVA_DB.Create(&rec).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

// GetCaActiveSm4Key 取启用密钥（解密后的 16 字节）
func (s *CaService) GetCaActiveSm4Key(keyName string) ([]byte, error) {
	var rec model.CaKeyRotation
	if err := global.GVA_DB.Where("key_name = ? AND status = ?", keyName, "启用").First(&rec).Error; err != nil {
		return nil, fmt.Errorf("[%d]密钥不存在或已停用: %s", ErrCodeCaSm4NotFound, keyName)
	}
	plain, err := crypto.Decrypt(rec.SecretEnc)
	if err != nil {
		return nil, fmt.Errorf("[%d]密钥解密失败: %v", ErrCodeCaSm4Param, err)
	}
	return []byte(plain), nil
}

// GetCaKeyRotations 轮换表列表（密钥不回显）
func (s *CaService) GetCaKeyRotations() ([]model.CaKeyRotation, error) {
	var list []model.CaKeyRotation
	err := global.GVA_DB.Order("id ASC").Limit(100).Find(&list).Error
	return list, err
}

// EncryptCaData 用指定密钥 SM4 加密（字段级加密适配）
func (s *CaService) EncryptCaData(keyName string, plaintext []byte) (string, error) {
	key, err := s.GetCaActiveSm4Key(keyName)
	if err != nil {
		return "", err
	}
	return SM4EncryptECB(key, plaintext)
}

// DecryptCaData 用指定密钥 SM4 解密
func (s *CaService) DecryptCaData(keyName string, hexText string) ([]byte, error) {
	key, err := s.GetCaActiveSm4Key(keyName)
	if err != nil {
		return nil, err
	}
	return SM4DecryptECB(key, hexText)
}
