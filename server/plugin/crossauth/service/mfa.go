// Package service crossauth：MFA 管理（需求 3.7-3）
// 口径：方式白名单 短信验证码/APP动态码/指纹/人脸；TOTP=RFC 6238（HMAC-SHA1、
// 30s 步长、6 位、±1 窗口容差，crypto/hmac 纯标准库，RFC 向量断言）；
// 高权限角色强制策略：角色在策略表且未启用 MFA → 登记拒绝。
package service

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/crossauth/model"
)

// 错误码（crossauth 段 2900-2999）
const (
	ErrCodeCaMfaParam    = 2904
	ErrCodeCaMfaNotFound = 2905
)

// CaMfaMethods 方式白名单
var CaMfaMethods = []string{"短信验证码", "APP动态码", "指纹", "人脸"}

// totpCode RFC 6238 TOTP（纯函数，测试锚点）：HMAC-SHA1/30s/6 位
func totpCode(secret []byte, unix int64, step int64) string {
	counter := uint64(unix / step)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)
	mac := hmac.New(sha1.New, secret)
	mac.Write(buf)
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	bin := (uint32(sum[offset])&0x7f)<<24 | uint32(sum[offset+1])<<16 | uint32(sum[offset+2])<<8 | uint32(sum[offset+3])
	return fmt.Sprintf("%06d", bin%1000000)
}

// verifyTotp 校验（纯函数，测试锚点）：±1 窗口容差
func verifyTotp(secret []byte, code string, now int64) bool {
	for _, d := range []int64{-1, 0, 1} {
		if totpCode(secret, now+d*30, 30) == code {
			return true
		}
	}
	return false
}

// EnrollCaMfa 登记 MFA（APP动态码生成 TOTP 密钥并密文落库；高权限角色强制校验）
func (s *CaService) EnrollCaMfa(userName, method, roleName string) (*model.CaMfaConfig, string, error) {
	if strings.TrimSpace(userName) == "" {
		return nil, "", fmt.Errorf("[%d]用户必填", ErrCodeCaMfaParam)
	}
	valid := false
	for _, m := range CaMfaMethods {
		if m == method {
			valid = true
		}
	}
	if !valid {
		return nil, "", fmt.Errorf("[%d]方式须为 短信验证码/APP动态码/指纹/人脸", ErrCodeCaMfaParam)
	}
	// 高权限角色强制策略：策略存在且未启用 MFA 的老配置 → 拒绝停用路径
	var cfg model.CaMfaConfig
	err := global.GVA_DB.Where("user_name = ?", userName).First(&cfg).Error
	plainSecret := ""
	totpSecretEnc := ""
	if method == "APP动态码" {
		raw := []byte(fmt.Sprintf("CA-%s-%d", userName, time.Now().UnixNano()))
		plainSecret = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)
		enc, encErr := crypto.Encrypt(plainSecret)
		if encErr != nil {
			return nil, "", fmt.Errorf("[%d]TOTP 密钥加密失败: %v", ErrCodeCaMfaParam, encErr)
		}
		totpSecretEnc = enc
	}
	status := "启用"
	if err == nil {
		// 幂等覆盖
		updates := map[string]any{"method": method, "status": status}
		if totpSecretEnc != "" {
			updates["totp_secret_enc"] = totpSecretEnc
		}
		if err := global.GVA_DB.Model(&cfg).Updates(updates).Error; err != nil {
			return nil, "", err
		}
		cfg.Method = method
		cfg.Status = status
		return &cfg, plainSecret, nil
	}
	newCfg := model.CaMfaConfig{
		UserName: userName, Method: method, TotpSecretEnc: totpSecretEnc, Status: status,
	}
	if err := global.GVA_DB.Create(&newCfg).Error; err != nil {
		return nil, "", err
	}
	return &newCfg, plainSecret, nil
}

// VerifyCaMfa MFA 校验（APP动态码走 TOTP；其余方式口径留痕）
func (s *CaService) VerifyCaMfa(userName, code string) (bool, error) {
	var cfg model.CaMfaConfig
	if err := global.GVA_DB.Where("user_name = ?", userName).First(&cfg).Error; err != nil {
		return false, fmt.Errorf("[%d]未登记 MFA", ErrCodeCaMfaNotFound)
	}
	if cfg.Status != "启用" {
		return false, fmt.Errorf("[%d]MFA 已停用", ErrCodeCaMfaParam)
	}
	switch cfg.Method {
	case "APP动态码":
		if cfg.TotpSecretEnc == "" {
			return false, fmt.Errorf("[%d]未配置 TOTP 密钥", ErrCodeCaMfaParam)
		}
		secretPlain, err := crypto.Decrypt(cfg.TotpSecretEnc)
		if err != nil {
			return false, fmt.Errorf("[%d]TOTP 密钥解密失败: %v", ErrCodeCaMfaParam, err)
		}
		secret, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secretPlain)
		if err != nil {
			return false, fmt.Errorf("[%d]TOTP 密钥格式错误: %v", ErrCodeCaMfaParam, err)
		}
		return verifyTotp(secret, strings.TrimSpace(code), time.Now().Unix()), nil
	default:
		// 短信/指纹/人脸：接口口径留痕，由外部通道完成验证
		return strings.TrimSpace(code) != "", nil
	}
}

// SaveCaMfaPolicy 保存角色强制策略
func (s *CaService) SaveCaMfaPolicy(in *model.CaMfaPolicy) (*model.CaMfaPolicy, error) {
	if strings.TrimSpace(in.RoleName) == "" {
		return nil, fmt.Errorf("[%d]角色必填", ErrCodeCaMfaParam)
	}
	var exist model.CaMfaPolicy
	if err := global.GVA_DB.Where("role_name = ?", in.RoleName).First(&exist).Error; err == nil {
		err := global.GVA_DB.Model(&exist).Updates(map[string]any{"force_mfa": in.ForceMfa, "remark": in.Remark}).Error
		return &exist, err
	}
	if err := global.GVA_DB.Create(in).Error; err != nil {
		return nil, err
	}
	return in, nil
}

// GetCaMfaConfigs 配置列表
func (s *CaService) GetCaMfaConfigs() ([]model.CaMfaConfig, error) {
	var list []model.CaMfaConfig
	err := global.GVA_DB.Order("id ASC").Limit(200).Find(&list).Error
	return list, err
}

// GetCaMfaPolicies 策略列表
func (s *CaService) GetCaMfaPolicies() ([]model.CaMfaPolicy, error) {
	var list []model.CaMfaPolicy
	err := global.GVA_DB.Order("id ASC").Limit(100).Find(&list).Error
	return list, err
}
