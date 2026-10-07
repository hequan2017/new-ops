package service

import (
	"fmt"

	"github.com/hequan2017/new-ops/server/global"
	"github.com/hequan2017/new-ops/server/plugin/asset/crypto"
	"github.com/hequan2017/new-ops/server/plugin/asset/model"
	"gorm.io/gorm"
)

// asset 插件错误码补充（凭据）
const (
	ErrCodeCredNameRequired = 1014
	ErrCodeCredTypeInvalid  = 1015
	ErrCodeCredSecretEmpty  = 1016
	ErrCodeCredInUse        = 1017
)

// CredCredentialService 凭据保险库服务
type CredCredentialService struct{}

// validateCredential 凭据校验（纯逻辑）
func validateCredential(c *model.CredCredential, secret string) error {
	if c == nil || c.Name == "" {
		return newServiceErr(ErrCodeCredNameRequired, "凭据名称不能为空")
	}
	if !model.ValidCredTypes()[c.Type] {
		return newServiceErr(ErrCodeCredTypeInvalid, fmt.Sprintf("非法的凭据类型: %s", c.Type))
	}
	if secret == "" {
		return newServiceErr(ErrCodeCredSecretEmpty, "凭据内容不能为空")
	}
	return nil
}

// CreateCredential 创建凭据（加密落库，只存密文与末4位）
func (s *CredCredentialService) CreateCredential(c *model.CredCredential, secret string) error {
	if err := validateCredential(c, secret); err != nil {
		return err
	}
	cipherText, err := crypto.Encrypt(secret)
	if err != nil {
		return err
	}
	c.Cipher = cipherText
	c.SecretLast4 = crypto.Last4(secret)
	return global.GVA_DB.Create(c).Error
}

// UpdateCredential 更新凭据（secret 为空则保留原密文，仅改名称/备注/用户名）
func (s *CredCredentialService) UpdateCredential(c *model.CredCredential, secret string) error {
	if c == nil || c.ID == 0 || c.Name == "" {
		return newServiceErr(ErrCodeCredNameRequired, "凭据ID与名称不能为空")
	}
	if c.Type != "" && !model.ValidCredTypes()[c.Type] {
		return newServiceErr(ErrCodeCredTypeInvalid, fmt.Sprintf("非法的凭据类型: %s", c.Type))
	}
	updates := map[string]any{
		"name":     c.Name,
		"username": c.Username,
		"remark":   c.Remark,
	}
	if secret != "" {
		cipherText, err := crypto.Encrypt(secret)
		if err != nil {
			return err
		}
		updates["cipher"] = cipherText
		updates["secret_last4"] = crypto.Last4(secret)
	}
	return global.GVA_DB.Model(&model.CredCredential{}).Where("id = ?", c.ID).Updates(updates).Error
}

// DeleteCredential 删除凭据（被引用时拒绝）
func (s *CredCredentialService) DeleteCredential(id uint) error {
	var cred model.CredCredential
	if err := global.GVA_DB.First(&cred, id).Error; err != nil {
		return err
	}
	if cred.RefCount > 0 {
		return newServiceErr(ErrCodeCredInUse, fmt.Sprintf("凭据「%s」被引用 %d 次，请先解除引用", cred.Name, cred.RefCount))
	}
	return global.GVA_DB.Delete(&cred).Error
}

// GetCredentialList 凭据列表（密文与明文均不返回，模型 Cipher 已 json:"-"）
func (s *CredCredentialService) GetCredentialList(keyword string) (list []*model.CredCredential, err error) {
	db := global.GVA_DB.Model(&model.CredCredential{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR username LIKE ? OR remark LIKE ?", kw, kw, kw)
	}
	err = db.Order("id DESC").Find(&list).Error
	return list, err
}

// GetPlaintext 取凭据明文（仅服务端内部使用：SSH 采集/容器接入等；不经过 HTTP 响应）
func (s *CredCredentialService) GetPlaintext(id uint) (plaintext, credType, username string, err error) {
	var cred model.CredCredential
	if err = global.GVA_DB.First(&cred, id).Error; err != nil {
		return "", "", "", err
	}
	plaintext, err = crypto.Decrypt(cred.Cipher)
	return plaintext, cred.Type, cred.Username, err
}

// IncrRefCount / DecrRefCount 引用计数维护（引用方创建/删除绑定关系时调用）
func (s *CredCredentialService) IncrRefCount(tx *gorm.DB, id uint) error {
	return tx.Model(&model.CredCredential{}).Where("id = ?", id).
		UpdateColumn("ref_count", gorm.Expr("ref_count + 1")).Error
}

func (s *CredCredentialService) DecrRefCount(tx *gorm.DB, id uint) error {
	return tx.Model(&model.CredCredential{}).Where("id = ? AND ref_count > 0", id).
		UpdateColumn("ref_count", gorm.Expr("ref_count - 1")).Error
}
